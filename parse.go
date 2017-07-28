package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Prefixes for all relevant types of outputs from the error log.
const (
	// Lines prepending test output.
	commitHashLine string = "At commit:"

	// Lines giving test results.
	runTest    string = "=== RUN"
	skipTest   string = "--- SKIP:"
	failedTest string = "--- FAIL:"
	passedTest string = "--- PASS:"

	// Lines giving package results.
	packageFail   string = "FAIL	github.com/"
	packagePass   string = "ok  	github.com/"
	packagePrefix string = "github.com/NebulousLabs/Sia/"

	//Reference time formatting for dateTimes.
	referenceTime string = "2006-01-02-15:04:05"
)

// ReadFile reads the file with the given name and returns a slice of string,
// one for each line of the file.
func ReadFile(name string) []string {
	var output []string
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		next := scanner.Text()
		output = append(output, next)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return output
}

// ParseErrorLog parses the given file and creates a result object using the
// information contained in the file.
func ParseErrorLog(name string) *Result {
	lines := ReadFile(name)

	dateTimeStr := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(name), ".log"), "error-")
	dateTime, err := time.Parse(referenceTime, dateTimeStr)
	if err != nil {
		fmt.Println(err)
	}

	// The other Result fields:
	var commitHash string
	var testResults []*TestResult
	var packageResults []*PackageResult

	testsStarted := make(map[string]struct{})
	for i := 0; i < len(lines); i++ {
		switch {
		case strings.HasPrefix(lines[i], commitHashLine):
			// Store the commit hash of the code run by this test.
			commitHash = strings.TrimSpace(lines[i+1])
			i++ //Skip the line with the hash we just stored.

		case strings.HasPrefix(lines[i], runTest):
			// Store the name of this test.
			testName := strings.TrimSpace(strings.TrimPrefix(lines[i], runTest))
			testsStarted[testName] = struct{}{}

		case strings.HasPrefix(lines[i], skipTest):
			skippedTest := strings.Split(strings.TrimSpace(strings.TrimPrefix(lines[i], skipTest)), " ")

			// Add all lines that contain information about this test.
			for j := 1; j < len(lines); j++ {
				nextLine := lines[i+j]
				if strings.HasPrefix(nextLine, "\t") {
					skippedTest = append(skippedTest, nextLine)
				} else {
					i += j // Account for the lines appended.
					break
				}
			}

			r := handleSkippedTest(skippedTest)
			// Add the result to the slice of results and remove the test name
			// from the set so that it doesn't get handled twice.
			testResults = append(testResults, r)
			delete(testsStarted, r.name)

		case strings.HasPrefix(lines[i], failedTest):
			fail := strings.Split(strings.TrimSpace(strings.TrimPrefix(lines[i], failedTest)), " ")

			// Add all lines that contain information about this test.
			for j := 1; j < len(lines); j++ {
				nextLine := lines[i+j]
				if strings.HasPrefix(nextLine, "\t") {
					fail = append(fail, nextLine)
				} else {
					i += j // Account for the lines appended.
					break
				}
			}

			r := handleFailedTest(fail)
			// Add the result to the slice of results and remove the test name
			// from the set so that it doesn't get handled twice.
			testResults = append(testResults, r)
			delete(testsStarted, r.name)

		case strings.HasPrefix(lines[i], passedTest):
			pass := strings.Split(strings.TrimSpace(strings.TrimPrefix(lines[i], passedTest)), " ")
			durStr := strings.TrimPrefix(strings.TrimSuffix(pass[1], ")"), "(") // Remove surrounding parentheses.
			dur, err := time.ParseDuration(durStr)
			if err != nil {
				fmt.Println(err)
			}
			r := &TestResult{
				name:     pass[0],
				result:   Status(PASSED),
				output:   "",
				duration: dur,
			}

			testResults = append(testResults, r)
			delete(testsStarted, r.name)

		case strings.HasPrefix(lines[i], packageFail):
			fail := strings.Split(strings.TrimSpace(strings.TrimPrefix(lines[i], "FAIL")), "\t")
			packageName := strings.TrimPrefix(fail[0], packagePrefix)
			packageDur, err := time.ParseDuration(fail[1])
			if err != nil {
				fmt.Println(err)
			}

			mr := &PackageResult{
				name:     packageName,
				result:   Status(FAILED),
				duration: packageDur,
			}
			packageResults = append(packageResults, mr)

		case strings.HasPrefix(lines[i], packagePass):
			pass := strings.Split(strings.TrimSpace(strings.TrimPrefix(lines[i], "ok")), "\t")
			packageName := strings.TrimPrefix(pass[0], packagePrefix)
			packageDur, err := time.ParseDuration(pass[1])
			if err != nil {
				fmt.Println(err)
			}

			mr := &PackageResult{
				name:     packageName,
				result:   Status(PASSED),
				duration: packageDur,
			}
			packageResults = append(packageResults, mr)

		default:
		}
	}

	// Add all tests that were started and not heard back from as 'UNDETERMINED' tests.
	for t := range testsStarted {
		r := &TestResult{
			name:     t,
			result:   Status(UNDETERMINED),
			output:   "",
			duration: 0,
		}
		testResults = append(testResults, r)
	}

	return &Result{
		commitHash:     commitHash,
		dateTime:       dateTime,
		testResults:    testResults,
		packageResults: packageResults,
	}
}

// handleSkippedTest creates a testResult object from a slice of strings in
// which the first element is the name of the test, the second element is the
// duration of the test in the form "(0.00s)", and the third(last) element is
// the output of the test.
func handleSkippedTest(testOutput []string) *TestResult {
	// Get the duration of the test.
	durStr := testOutput[1]
	durStr = strings.TrimPrefix(strings.TrimSuffix(durStr, ")"), "(") // Remove surrounding parentheses.
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		fmt.Println(err)
	}

	// A skipped test can have 0 or more lines of output.
	var out string
	if len(testOutput) < 3 {
		out = ""
	} else {
		for i := 2; i < len(testOutput); i++ {
			out = out + testOutput[i]
		}
	}

	// Get the name of the test and other fields.
	testName := testOutput[0]
	r := &TestResult{
		name:     testName,
		result:   Status(SKIPPED),
		output:   out,
		duration: dur,
	}

	return r
}

// handleFailedTest creates a testResult object from a slice of strings in
// which the first element is the name of the test, the second element is the
// duration of the test in the form "(0.00s)", and the third(last) element is
// the output of the test.
func handleFailedTest(testOutput []string) *TestResult {
	// Get the duration of the test.
	durStr := testOutput[1]
	durStr = strings.TrimPrefix(strings.TrimSuffix(durStr, ")"), "(") // Remove surrounding parentheses.
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		fmt.Println(err)
	}

	// A failed test can have 0 or more lines of output.
	var out string
	if len(testOutput) < 3 {
		out = ""
	} else {
		for i := 2; i < len(testOutput); i++ {
			out = out + testOutput[i]
		}
	}

	// Get the name of the test and other fields.
	testName := testOutput[0]
	r := &TestResult{
		name:     testName,
		result:   Status(FAILED),
		output:   out,
		duration: dur,
	}

	return r
}
