package main

import (
	"database/sql"
	"net/http"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/bmizerany/pat"
)

var musicDB *sql.DB

type Track struct {
	Id	int64	`json:"id"`
	Artist	string	`json:"artist"`
	Album	string	`json:"album"`
	Name	string	`json:"name"`
	Uri	string	`json:"uri"`
	Added	string	`json:"added"`
}

type Tracks []Track

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

func getById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")

	fmt.Println("ID:", id)

	statement, stmErr := musicDB.Prepare("SELECT * FROM music WHERE id = ?")
	checkErr(stmErr)

	rows, rowErr := statement.Query(id)
	checkErr(rowErr)

	var track Track
	for rows.Next() {
		rows.Scan(&track.Id, &track.Artist, &track.Album, &track.Name, &track.Uri, &track.Added)
	}

	fmt.Println(track)

	jsonB, errMarshal := json.Marshal(track)

	checkErr(errMarshal)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(jsonB))
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
	r.Get("/tracks/:id", http.HandlerFunc(getById))

	http.Handle("/", r)
	
	httpErr := http.ListenAndServe(":6789", nil)
	checkErr(httpErr)
}
