package main

import (
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func (env *Environment) InsertLogToDB(filename string) {
	results := ParseErrorLog(filename)

	testStmt, err := env.db.Prepare("INSERT tests SET commitHash=?,dateTime=?,name=?,result=?,output=?,duration=?")
	if err != nil {
		log.Fatal("Error preparing test insert statement: ", err)
	}
	packageStmt, err := env.db.Prepare("INSERT packages SET commitHash=?,dateTime=?,name=?,result=?,duration=?")
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

func (env *Environment) InsertLogsFromDirectory(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading directory: ", err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		env.InsertLogToDB(dir + f.Name())
	}
}

func (env *Environment) mostRecentCommitHash() string {
	var hash string
	err := env.db.QueryRow("select commitHash from tests order by datetime desc limit 1;").Scan(&hash)
	if err != nil {
		log.Fatal("Error getting commit hash: ", err)
	}
	return hash
}
