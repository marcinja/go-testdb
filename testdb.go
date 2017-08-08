package main

import (
	"database/sql"
	"flag"
	"fmt"
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
		env.DailyUpdate("fileName")
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

func (env *Environment) DailyUpdate(fileName string) {
	diffs := env.performanceDiffsFromLastWeek()
	// TODO: write to a file (nicely)

	failedTests := env.failedTestsFromLastDay()
	// TODO: write to a file (nicely)

	panics := env.panicsFromLastDay()
	// TODO: turn panics into slice of file names (i.e. the logs contained panic information)

	println(len(diffs))
	println(len(failedTests))
	println(len(panics))

}
