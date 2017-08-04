package main

import (
	"database/sql"
	"log"
	"time"
)

/*

GOAL:

select name, dateTime, output from tests where datetime between date_sub(now(), INTERVAL 1 day) and now() and result='FAILED';

TODO: Panic catcher will also return file name

TODO: should this have a NULL safe sql dateTime type?

*?


*/

// TestResult represents the information given from a single test fail.
type failResult struct {
	commitHash string
	dateTime   time.Time
	name       string
	result     Status
	output     string
	duration   time.Duration
}

func failedTestsFromLastWeek(db *sql.DB) []*failResult {
	rows, err := db.Query(" select commitHash, dateTime, name, output, duration from tests where datetime between date_sub(now(), INTERVAL 1 day) and now() and result='FAILED';")
	if err != nil {
		log.Fatal("Error selecting failed results: ", err)
	}
	var results []*failResult
	defer rows.Close()
	for rows.Next() {
		// failResult fields:
		var (
			hash     sql.NullString
			dateTime time.Time
			name     sql.NullString
			output   sql.NullString
			duration time.Duration
		)
		err := rows.Scan(&hash, &dateTime, &name, &output, &duration)
		if err != nil {
			log.Fatal(err)
		}

		// SQL can return NULL types for strings. If we get a NULL string we
		// just keep it as an empty string.
		var safeHash string
		var safeName string
		var safeOutput string
		if hash.Valid {
			safeHash = hash.String
		}
		if name.Valid {
			safeName = name.String
		}
		if output.Valid {
			safeOutput = output.String
		}

		fr := &failResult{
			commitHash: safeHash,
			dateTime:   dateTime,
			name:       safeName,
			result:     Status(FAILED),
			output:     safeOutput,
			duration:   duration,
		}

		results = append(results, fr)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return results
}
