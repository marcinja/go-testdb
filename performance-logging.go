package main

import (
	"database/sql"
	"log"
	"math"
)

const (
	avgFromCommitHash string = "select AVG(duration) from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() and name = ? and commitHash = ? and result='PASSED';"
	avgFromLastWeek   string = "select AVG(duration) from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() and name = ? and result='PASSED' and commitHash != ?;"
)

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
func averageDurationFromCommitHash(stmt *sql.Stmt, testName string, latestCommit string) (float64, bool) {
	var avg sql.NullFloat64

	rows, err := stmt.Query(testName, latestCommit)
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
func averageDurationFromLastWeek(stmt *sql.Stmt, testName string, hash string) (float64, bool) {
	var avg sql.NullFloat64
	rows, err := stmt.Query(testName, hash)
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
func (e *Environment) performanceDiffsFromLastWeek() []*performanceDiff {
	latestCommit := e.mostRecentCommitHash()
	testNames := e.testNamesFromLastWeek()

	avgFromLastWeekStmt, err := e.db.Prepare(avgFromLastWeek)
	if err != nil {
		log.Fatal(err)
	}
	defer avgFromLastWeekStmt.Close()

	avgFromCommitHashStmt, err := e.db.Prepare(avgFromCommitHash)
	if err != nil {
		log.Fatal(err)
	}
	defer avgFromCommitHashStmt.Close()

	var diffs []*performanceDiff
	for _, name := range testNames {
		avgFromResults, ok := averageDurationFromLastWeek(avgFromLastWeekStmt, name, latestCommit)
		if !ok {
			continue
		}
		avgFromRecentResults, ok := averageDurationFromCommitHash(avgFromCommitHashStmt, name, latestCommit)
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
func (env *Environment) testNamesFromLastWeek() []string {
	const nameQuery string = "select name from tests where datetime between date_sub(now(), INTERVAL 1 WEEK) and now() group by name;"

	var results []string
	// Make query to db.
	rows, err := env.db.Query(nameQuery)
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
