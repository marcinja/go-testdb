package main

import (
	"database/sql"
	"log"
	"time"
)

/*

use sql query to get list of test names "group by name"

select name from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() and result='FAILED' group by name;

then: make one query for each test keeping track of name, duration


get test results from most recent commit hash

check average of those, compare to avg over last week

*/

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

func resultsFromLastWeek(testName string, db *sql.DB) []*TestResult {
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
		log.Println(duration)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return results
}
