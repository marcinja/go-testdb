package main

import (
	"database/sql"
	"log"
	"math"
	"time"
)

// type performanceDiff struct {
// 	name                string
// 	lastWeeksAvgPerf    float64
// 	latestCommitAvgPerf float64
// }

type performanceDiff struct {
	name              string
	performanceChange float64
}

// averageDuration gives the average duraction of the given slice of test
// results.
func averageDuration(results []*TestResult) float64 {
	// This function is reaaaally slow. It's much better to ask MySql to do the
	// averaging for you!
	var totalSum float64
	for _, tr := range results {
		totalSum += tr.duration.Seconds()
	}

	if len(results) == 0 {
		return 0
	} else {
		avg := totalSum / float64(len(results))
		if avg < 0.1 {
			return 0
		} else {
			return avg
		}
	}
}

// averageDurationFromCommitHash returns the average duration for a given test
// at a specific commit hash. Returns true if the result from MySql is not NULL.
func averageDurationFromCommitHash(testName string, latestCommit string, db *sql.DB) (float64, bool) {
	var avg sql.NullFloat64

	rows, err := db.Query("select AVG(duration) from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() and name = ? and commitHash = ? and result='PASSED';", testName, latestCommit)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&avg)
		if err != nil {
			panic(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	if avg.Valid {
		return avg.Float64, true
	} else {
		return 0, false
	}

}

// averageDurationFromLastWeek returns the average duration for a given test for
// the last week, ignoring results from the most recent commit. Returns true if
// the result from MySql is not NULL.
func averageDurationFromLastWeek(testName string, hash string, db *sql.DB) (float64, bool) {
	var avg sql.NullFloat64
	rows, err := db.Query("select AVG(duration) from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() and name = ? and result='PASSED' and commitHash != ?;", testName, hash)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&avg)
		if err != nil {
			panic(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	if avg.Valid {
		return avg.Float64, true
	} else {
		return 0, false
	}
}

// performanceDiffsFromLastWeek returns a slice of performance diffs which
// summarize performance changes for tests greater than 5 seconds in length and
// which saw a 20% or greater change in performance over the last week compared
// to the most recent commit.
func performanceDiffsFromLastWeek(db *sql.DB) []*performanceDiff {
	latestCommit := mostRecentCommitHash(db)
	testNames := testNamesFromLastWeek(db)

	var diffs []*performanceDiff
	for _, name := range testNames {
		avgFromResults, ok := averageDurationFromLastWeek(name, latestCommit, db)
		if !ok {
			continue
		}
		avgFromRecentResults, ok := averageDurationFromCommitHash(name, latestCommit, db)
		if !ok {
			continue
		}

		// Short tests are ignored.
		if avgFromResults < 5.0 && avgFromRecentResults < 5.0 {
			continue
		}

		performanceDifference := avgFromRecentResults - avgFromResults
		//Add to diff if there is more than a 20% difference in performance.
		if math.Abs(performanceDifference) >= avgFromResults*0.2 && math.Abs(performanceDifference) >= 2.5 {
			println("\nDIFF: ", performanceDifference)
			println(name)
			diff := &performanceDiff{
				name:              name,
				performanceChange: performanceDifference,
			}
			diffs = append(diffs, diff)
		}
	}
	return diffs
}

// testNamesFromLastWeek returns a slice containing the names of every
// individual test that was run in the last week.
func testNamesFromLastWeek(db *sql.DB) []string {
	const nameQuery string = "select name from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() group by name;"

	var results []string
	// Make query to db.
	rows, err := db.Query(nameQuery)
	if err != nil {
		panic(err)
	}

	// Collect results of query.
	defer rows.Close()
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return results
}

func passingResultsFromLastWeek(testName string, db *sql.DB) []*TestResult {
	rows, err := db.Query("select output, duration from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() and result='PASSED' and name=?;", testName)
	if err != nil {
		log.Fatal(err)
	}

	var results []*TestResult
	defer rows.Close()
	for rows.Next() {
		// TestResult fields
		var (
			output   string
			duration time.Duration
		)
		err := rows.Scan(&output, &duration)
		if err != nil {
			log.Fatal(err)
		}
		tr := &TestResult{
			name:     testName,
			result:   Status(PASSED),
			output:   output,
			duration: duration,
		}

		results = append(results, tr)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return results
}

func resultsFromCommitHash(hash string, testName string, db *sql.DB) []*TestResult {
	rows, err := db.Query("select output, duration from tests where name = ? and commitHash = ? and result='PASSED'", testName, hash)
	if err != nil {
		log.Fatal("Error selecting commit hash results: ", err)
	}
	var results []*TestResult
	defer rows.Close()
	for rows.Next() {
		// TestResult fields
		var (
			output   string
			duration time.Duration
		)
		err := rows.Scan(&output, &duration)
		if err != nil {
			log.Fatal(err)
		}
		tr := &TestResult{
			name:     testName,
			result:   Status(PASSED),
			output:   output,
			duration: duration,
		}

		results = append(results, tr)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return results
}
