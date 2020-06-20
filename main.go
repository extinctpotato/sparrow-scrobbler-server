package main

import (
	"database/sql"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/bmizerany/pat"
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

func insert(w http.ResponseWriter, r *http.Request) {
	artist, artistOk := r.URL.Query()["artist"]
	album, albumOk := r.URL.Query()["album"]
	name, nameOk := r.URL.Query()["name"]

	_ = artistOk
	_ = albumOk
	_ = nameOk

	addRecord(artist[0], album[0], name[0], "")
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

	r := pat.New()
	r.Post("/tracks", http.HandlerFunc(insert))

	http.Handle("/", r)
	
	httpErr := http.ListenAndServe(":6789", nil)
	checkErr(httpErr)
}
