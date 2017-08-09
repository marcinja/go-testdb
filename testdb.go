package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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
		if *filePtr == "" {
			fmt.Printf("Run this command with the '-file FILENAME' flag.")
		}
		env.DailyUpdate(*filePtr)
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

func (env *Environment) DailyUpdate(filepath string) {
	diffs := env.performanceDiffsFromLastWeek()
	failedTests := env.failedTestsFromLastDay()
	panics := env.panicsFromLastDay()

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatal("Error creating file for update: ", err)
	}
	defer f.Close()

	f.WriteString("Found " + strconv.Itoa(len(panics)) + " panics in tests.\n")
	for _, t := range panics {
		f.WriteString(t.String() + "\n")
	}

	f.WriteString("Found " + strconv.Itoa(len(failedTests)) + " tests that failed.\n")
	for _, test := range failedTests {
		f.WriteString("Name: " + test.name + "\n")
		f.WriteString("Commit Hash: " + test.commitHash + "\n")
		f.WriteString("Datetime: " + test.dateTime.String() + "\n")
		f.WriteString("Duration: " + test.duration.String() + " seconds.\n")
		f.WriteString("Output: " + test.output + "\n")
	}

	f.WriteString("Found " + strconv.Itoa(len(diffs)) + " tests whose performance changed by more than 20%.\n")
	for _, diff := range diffs {
		f.WriteString("Name: " + diff.name + "\n")
		f.WriteString("Name: " + strconv.FormatFloat(diff.performanceChange, 'f', -1, 64) + "\n")
	}
	f.Sync()
	fmt.Printf("Daily update written to file succesfully.")
}
