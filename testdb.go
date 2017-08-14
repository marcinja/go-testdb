package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

/*

TODO: Clean up repetitious code. A lot of functions do "select from ..."

TODO: Find average durations concurrently

TODO: Decide on proper way to output results.

TODO: Send results in email, once daily.

*/

type Environment struct {
	db *sql.DB
}

var dbInfoFile string

func main() {
	dirPtr := flag.String("dir", "", "directory path")
	filePtr := flag.String("file", "", "file path")
	dbInfoPtr := flag.String("dbinfo", "db-info.txt", "file in which db information is contained")
	updatePtr := flag.Bool("getUpdate", false, "receive an informed db update at the stated file path")

	emailPtr := flag.String("email", "", "the email that will recieve the update")
	namePtr := flag.String("name", "", "the name of the person that will recieve the update email")
	flag.Parse()

	dbInfoFile = *dbInfoPtr // Set directory for db info.
	dbInfo := ReadFile(dbInfoFile)[0]
	db, err := sql.Open("mysql",
		dbInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	env := &Environment{
		db,
	}

	if *updatePtr {
		if *filePtr != "" {
			env.DailyUpdateToFile(*filePtr)
			return
		}

		if *emailPtr == "" && *namePtr == "" {
			fmt.Printf("Run this command with the '-file FILENAME' flag, or with -email and -name flags.")
			return
		}

		subj, body := env.DailyUpdate()
		email(*emailPtr, *namePtr, subj, body)

		return
	}

	if *dirPtr != "" {
		env.InsertLogsFromDirectory(*dirPtr)
	} else if *filePtr != "" {
		env.InsertLogToDB(*filePtr)
	} else {
		fmt.Printf("No directory or file path given.")
	}
}

// DailyUpdate gets panics, test failures, and performance changes from the last
// day and outputs two strings fit for email subject and body that describe
// these changes.
func (env *Environment) DailyUpdate() (subject string, body string) {
	diffs := env.performanceDiffsFromLastWeek()
	failedTests := env.failedTestsFromLastDay()
	panics := env.panicsFromLastDay()

	body += "Found " + strconv.Itoa(len(panics)) + " panics in tests.\n"
	for _, t := range panics {
		body += "\n" + t.Format(referenceTime) + "\n"
	}

	body += "\nFound " + strconv.Itoa(len(failedTests)) + " tests that failed.\n"
	for _, test := range failedTests {
		body += "\n\tName: " + test.name + "\n"
		body += "\tCommit Hash: " + test.commitHash + "\n"
		body += "\tDatetime: " + test.dateTime.Format(referenceTime) + "\n"
		body += "\tDuration: " + strings.TrimSuffix(test.duration.String(), "ns") + " seconds.\n"
		body += "\tOutput: " + test.output + "\n"
	}

	body += "\nFound " + strconv.Itoa(len(diffs)) + " tests whose performance changed by more than 20%.\n\n"
	for _, diff := range diffs {
		body += "\n\tName: " + diff.name + "\n"
		body += "\tName: " + strconv.FormatFloat(diff.performanceChange, 'f', -1, 64) + "\n"
	}
	subject = "CI Update: Found " + strconv.Itoa(len(panics)) + " panics, " + strconv.Itoa(len(failedTests)) + " test failures, " + strconv.Itoa(len(diffs)) + " performance changes"

	return subject, body
}

// DailyUpdateToFile performs a DailyUpdate and writes the result to a file at
// the given path.
func (env *Environment) DailyUpdateToFile(filepath string) {
	subj, body := env.DailyUpdate()

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal("Error creating file for update: ", err)
	}
	defer f.Close()

	f.WriteString(subj + "\n")
	f.WriteString(body)
	f.Sync()
	fmt.Printf("Daily update written to file succesfully.")
}

// email sends an email to the recipient email with the recipient name, subject,
// and body.
func email(recipientEmail, recipientName, subject, body string) error {
	emailAddr := os.Getenv("EMAIL_ADDR")
	emailPw := os.Getenv("EMAIL_PW")
	s, err := gomail.NewDialer("smtp.gmail.com", 587, emailAddr, emailPw).Dial()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailAddr)
	m.SetAddressHeader("To", recipientEmail, recipientName)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	return gomail.Send(s, m)
}
