package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var musicDB *sql.DB

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func addRecord(artist string, album string, name string, uri string) {
	statement, err := musicDB.Prepare(
		`INSERT INTO music (artist, album, name, uri, time)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	checkErr(err)
	statement.Exec(artist, album, name, uri)
}

func main() {
	db, err := sql.Open("sqlite3", "./tracks.db")
	defer db.Close()
	checkErr(err)
	musicDB = db

	statement, err := db.Prepare(
		`CREATE TABLE IF NOT EXISTS music 
		(id INTEGER PRIMARY KEY, artist TEXT, 
		album TEXT, name TEXT, uri TEXT, time CURRENT_TIMESTAMP)`)
	statement.Exec()

	addRecord("cool_artist", "cool_album", "cool_name", "uri://cooluri")
}
