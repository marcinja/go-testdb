package main

import (
	"database/sql"
	"io/ioutil"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	InsertLogsFromDirectory("/home/marcinja/error-logs/")
}

func InsertLogToDB(filename string, db *sql.DB) {
	results := ParseErrorLog(filename)
	dt := results.dateTime.Format(referenceTime)
	log.Println(dt)

	testStmt, err := db.Prepare("INSERT tests SET commitHash=?,dateTime=?,name=?,result=?,output=?,duration=?")
	if err != nil {
		panic(err)
	}
	moduleStmt, err := db.Prepare("INSERT modules SET commitHash=?,dateTime=?,name=?,result=?,duration=?")
	if err != nil {
		panic(err)
	}

	for _, t := range results.testResults {
		statusString := StatusStrings[int(t.result)]
		_, err := testStmt.Exec(results.commitHash, results.dateTime, t.name, statusString, t.output, int(t.duration.Seconds()))
		if err != nil {
			panic(err)
		}
	}

	for _, m := range results.moduleResults {
		statusString := StatusStrings[int(m.result)] // MySql expects a string type for its enum.
		_, err := moduleStmt.Exec(results.commitHash, results.dateTime, m.name, statusString, int(m.duration.Seconds()))
		if err != nil {
			panic(err)
		}
	}
}

func InsertLogsFromDirectory(dir string) {
	dbInfo := ReadFile("db-info.txt")[0]
	db, err := sql.Open("mysql",
		dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		InsertLogToDB(dir+f.Name(), db)
	}
}
