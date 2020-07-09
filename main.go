package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"encoding/json"
	"fmt"
	"strconv"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/bmizerany/pat"
)

var musicDB *sql.DB

var sClientId string
var sClientSecret string
var sCallbackUrl string

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

/* Pure SQLite-related functions */

func addRecord(artist string, album string, name string, uri string) int64 {
	statement, err := musicDB.Prepare(
		`INSERT INTO music (artist, album, name, uri, time)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	checkErr(err)

	result, resultErr := statement.Exec(artist, album, name, uri)
	checkErr(resultErr)

	id, idErr := result.LastInsertId()
	checkErr(idErr)

	return id
}

func idToJson(id string) string {
	statement, stmErr := musicDB.Prepare("SELECT * FROM music WHERE id = ?")
	checkErr(stmErr)

	rows, rowErr := statement.Query(id)
	checkErr(rowErr)

	var track Track
	for rows.Next() {
		rows.Scan(&track.Id, &track.Artist, &track.Album, &track.Name, &track.Uri, &track.Added)
	}

	jsonB, errMarshal := json.Marshal(track)
	checkErr(errMarshal)

	return string(jsonB)
}

func allRecords(page int) string {
	idOffset := 30*page
	statement, stmErr := musicDB.Prepare(
		`SELECT * FROM music
		WHERE id > ?
		ORDER BY id
		LIMIT 30`)
	checkErr(stmErr)

	rows, rowErr := statement.Query(idOffset)
	checkErr(rowErr)

	var tracks Tracks
	for rows.Next() {
		var track Track
		rows.Scan(&track.Id, &track.Artist, &track.Album, &track.Name, &track.Uri, &track.Added)
		tracks = append(tracks, track)
	}

	jsonB, errMarshal := json.Marshal(tracks)
	checkErr(errMarshal)

	return string(jsonB)
}

func setConf(key string, value string) {
	statement, stmErr := musicDB.Prepare("UPDATE conf SET value = ? WHERE key = ?")
	checkErr(stmErr)

	result, resultErr := statement.Exec(value, key)
	checkErr(resultErr)
	_ = result
}

func getConf(key string) string {
	statement, stmErr := musicDB.Prepare("SELECT value FROM conf WHERE key = ?")
	checkErr(stmErr)

	result, resultErr := statement.Query(key)
	checkErr(resultErr)

	var value string
	var iterErr error
	for result.Next() {
		iterErr = result.Scan(&value)
		checkErr(iterErr)
	}

	return value
}

/* REST-related functions */

func insert(w http.ResponseWriter, r *http.Request) {
	artist, artistOk := r.URL.Query()["artist"]
	album, albumOk := r.URL.Query()["album"]
	name, nameOk := r.URL.Query()["name"]

	_ = artistOk
	_ = albumOk
	_ = nameOk

	newId := addRecord(artist[0], album[0], name[0], "")
	serializedId := idToJson(strconv.FormatInt(newId, 10))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", serializedId)
}

func getById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")

	fmt.Println("ID:", id)

	serializedId := idToJson(id)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", serializedId)
}

func getAllPaged(w http.ResponseWriter, r *http.Request) {
	page, pageOk := r.URL.Query()["page"]
	_ = pageOk

	pageInt, pageIntErr := strconv.Atoi(page[0])
	checkErr(pageIntErr)

	pagedRecords := allRecords(pageInt)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", pagedRecords)
}

func spotifyAuthorize(w http.ResponseWriter, r *http.Request) {
	baseUrl, _ := url.Parse("https://accounts.spotify.com")
	baseUrl.Path += "authorize"

	params := url.Values{}
	params.Add("client_id", sClientId)
	params.Add("response_type", "code")
	params.Add("redirect_uri", sCallbackUrl)
	params.Add("scope", "user-read-recently-played")

	baseUrl.RawQuery = params.Encode()

	http.Redirect(w, r, baseUrl.String(), 301)
}

/* main */

func main() {
	db, err := sql.Open("sqlite3", "./tracks.db")
	defer db.Close()
	checkErr(err)
	musicDB = db

	sClientId = os.Getenv("SPOTIFY_CLIENT_ID")
	sClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	sCallbackUrl = os.Getenv("SPOTIFY_CALLBACK_URL")

	statementMusic, errMusic := db.Prepare(
		`CREATE TABLE IF NOT EXISTS music 
		(id INTEGER PRIMARY KEY, artist TEXT, 
		album TEXT, name TEXT, uri TEXT, time CURRENT_TIMESTAMP)`)
	checkErr(errMusic)
	statementMusic.Exec()

	statementConf, errConf := db.Prepare(
		`CREATE TABLE IF NOT EXISTS conf
		(key TEXT PRIMARY KEY, value TEXT)`)
	checkErr(errConf)
	statementConf.Exec()

	statementConfBlank, errConfBlank := db.Prepare(
		`INSERT OR IGNORE INTO conf (key, value) 
		VALUES ("ACCESS",""), ("REFRESH","")`)
	checkErr(errConfBlank)
	statementConfBlank.Exec()

	r := pat.New()
	r.Post("/api/tracks", http.HandlerFunc(insert))
	r.Get("/api/tracks", http.HandlerFunc(getAllPaged))
	r.Get("/api/tracks/:id", http.HandlerFunc(getById))
	r.Get("/api/sauth", http.HandlerFunc(spotifyAuthorize))

	http.Handle("/", r)

	fmt.Println("Client:", sClientId)
	fmt.Println("Secret:", sClientSecret)
	fmt.Println("Callback:", sCallbackUrl)

	httpErr := http.ListenAndServe(":6789", nil)
	checkErr(httpErr)
}
