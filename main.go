package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
)

var dbInfoFile string

func main2() {
	dirPtr := flag.String("dir", "", "directory path")
	filePtr := flag.String("file", "", "file path")
	dbInfoPtr := flag.String("dbinfo", "db-info.txt", "file in which db information is contained")
	flag.Parse()

	dbInfoFile = *dbInfoPtr // Set directory for db info.

	if *dirPtr != "" {
		InsertLogsFromDirectory(*dirPtr)
	} else if *filePtr != "" {
		InsertSingleLog(*filePtr)
	} else {
		fmt.Printf("No directory or file path given.")
	}

}

func main() {
	//dirPtr := flag.String("dir", "", "directory path")
	//filePtr := flag.String("file", "", "file path")
	dbInfoPtr := flag.String("dbinfo", "db-info.txt", "file in which db information is contained")
	flag.Parse()

	dbInfoFile = *dbInfoPtr // Set directory for db info.

	/*
		if *dirPtr != "" {
			InsertLogsFromDirectory(*dirPtr)
		} else if *filePtr != "" {
			InsertSingleLog(*filePtr)
		} else {
			fmt.Printf("No directory or file path given.")
		}
	*/
	dbInfo := ReadFile(dbInfoFile)[0]
	db, err := sql.Open("mysql",
		dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	names := testNamesFromLastWeek(db)

	for _, name := range names {
		r := resultsFromLastWeek(name, db)
		for _, res := range r {
			println(res.duration)
		}
	}

}
