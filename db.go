package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func InsertLogToDB(filename string, db *sql.DB) {
	results := ParseErrorLog(filename)

	testStmt, err := db.Prepare("INSERT tests SET commitHash=?,dateTime=?,name=?,result=?,output=?,duration=?")
	if err != nil {
		log.Fatal("Error preparing test insert statement: ", err)
	}
	packageStmt, err := db.Prepare("INSERT packages SET commitHash=?,dateTime=?,name=?,result=?,duration=?")
	if err != nil {
		log.Fatal("Error preparing package insert statement: ", err)
	}

	for _, t := range results.testResults {
		statusString := StatusStrings[int(t.result)]
		_, err := testStmt.Exec(results.commitHash, results.dateTime, t.name, statusString, t.output, int(t.duration.Seconds()))
		if err != nil {
			fmt.Println("Error inserting test result: ", err)
		}
	}

	for _, m := range results.packageResults {
		statusString := StatusStrings[int(m.result)] // MySql expects a string type for its enum.
		_, err := packageStmt.Exec(results.commitHash, results.dateTime, m.name, statusString, int(m.duration.Seconds()))
		if err != nil {
			fmt.Println("Error inserting package result: ", err)
		}
	}
}

func InsertLogsFromDirectory(dir string) {
	dbInfo := ReadFile(dbInfoFile)[0]
	db, err := sql.Open("mysql",
		dbInfo)
	if err != nil {
		log.Fatal("Error opening db: ", err)
	}
	defer db.Close()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading directory: ", err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		InsertLogToDB(dir+f.Name(), db)
	}
}

func InsertSingleLog(filename string) {
	dbInfo := ReadFile(dbInfoFile)[0]
	db, err := sql.Open("mysql",
		dbInfo)
	if err != nil {
		log.Fatal("Error reading file: ", err)
	}
	defer db.Close()

	InsertLogToDB(filename, db)
}

func mostRecentCommitHash(db *sql.DB) string {
	var hash string
	err := db.QueryRow("select commitHash from tests order by datetime desc limit 1;").Scan(&hash)
	if err != nil {
		log.Fatal("Error getting commit hash: ", err)
	}
	return hash
}
