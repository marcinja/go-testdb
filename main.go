package main

import (
	"flag"
	"fmt"
)

var dbInfoFile string

func main() {
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
