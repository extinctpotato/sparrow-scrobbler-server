package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var (
	musicDB       *sql.DB
	sClientId     string
	sClientSecret string
	sCallbackUrl  string
	sDbPath       string
)

const (
	ENV1 = "SPOTIFY_CLIENT_ID"
	ENV2 = "SPOTIFY_CLIENT_SECRET"
	ENV3 = "SPOTIFY_CALLBACK_URL"
	ENV4 = "SPOTIFY_DB_PATH"
)

type Tracks []Track

func checkErr(err error) {
	if err != nil {
		glog.Error(err)
	}
}

/* Pure SQLite-related functions */

func addRecord(p ...string) int64 {
	// p are ordered as follows: artist, album, name, uri, played_at

	statement, err := musicDB.Prepare(
		`INSERT INTO music (artist, album, name, uri, add_time, played_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?)`)
	checkErr(err)
	defer statement.Close()

	result, resultErr := statement.Exec(p[0], p[1], p[2], p[3], p[4])
	checkErr(resultErr)

	id, idErr := result.LastInsertId()
	checkErr(idErr)

	return id
}

func idToJson(id string) string {
	statement, stmErr := musicDB.Prepare("SELECT * FROM music WHERE id = ?")
	checkErr(stmErr)
	defer statement.Close()

	rows, rowErr := statement.Query(id)
	checkErr(rowErr)
	defer rows.Close()

	var track Track
	for rows.Next() {
		rows.Scan(&track.Id, &track.Artist, &track.Album, &track.Name, &track.Uri, &track.Added, &track.PlayedAt)
	}

	jsonB, errMarshal := json.Marshal(track)
	checkErr(errMarshal)

	return string(jsonB)
}

func allRecords(page int) string {
	maxIdStatement, _ := musicDB.Prepare("SELECT MAX(id) FROM music")
	defer maxIdStatement.Close()
	var maxId int
	maxIdStatement.QueryRow().Scan(&maxId)

	minPageId := maxId - 30*(page+1)
	maxPageId := minPageId + 30

	statement, _ := musicDB.Prepare(
		`SELECT * FROM music
		WHERE id > ? AND ID <= ?
		ORDER BY id DESC`)
	defer statement.Close()

	glog.Infof("Getting records from %d to %d.", minPageId, maxPageId)

	rows, rowErr := statement.Query(minPageId, maxPageId)
	checkErr(rowErr)
	defer rows.Close()

	var tracks Tracks
	for rows.Next() {
		var track Track
		rows.Scan(&track.Id, &track.Artist, &track.Album, &track.Name, &track.Uri, &track.Added, &track.PlayedAt)
		tracks = append(tracks, track)
	}

	jsonB, errMarshal := json.Marshal(tracks)
	checkErr(errMarshal)

	return string(jsonB)
}

func setConf(key string, value string) {
	statement, stmErr := musicDB.Prepare("UPDATE conf SET value = ? WHERE key = ?")
	defer statement.Close()
	checkErr(stmErr)

	result, resultErr := statement.Exec(value, key)
	checkErr(resultErr)
	_ = result
}

func getConf(key string) string {
	statement, stmErr := musicDB.Prepare("SELECT value FROM conf WHERE key = ?")
	defer statement.Close()
	checkErr(stmErr)

	result, resultErr := statement.Query(key)
	checkErr(resultErr)
	defer result.Close()

	var value string
	var iterErr error
	for result.Next() {
		iterErr = result.Scan(&value)
		checkErr(iterErr)
	}

	return value
}

/* REST-related functions */

func jsonError(w http.ResponseWriter, r *http.Request, c int, d string) {
	problem := Problem{
		Type:     "about:blank",
		Title:    http.StatusText(c),
		Status:   int64(c),
		Detail:   d,
		Instance: string(r.URL.Path),
	}

	data, err := json.Marshal(problem)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(c)
	w.Header().Set("Content-Type", "application/problem+json")
	fmt.Fprintf(w, string(data))
}

func insert(w http.ResponseWriter, r *http.Request) {
	artist, artistOk := r.URL.Query()["artist"]
	album, albumOk := r.URL.Query()["album"]
	name, nameOk := r.URL.Query()["name"]

	_ = artistOk
	_ = albumOk
	_ = nameOk

	newId := addRecord(artist[0], album[0], name[0], "", "")
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

	var pageInt int
	var pageIntErr error

	if len(page) > 0 {
		pageInt, pageIntErr = strconv.Atoi(page[0])
		checkErr(pageIntErr)
	} else {
		pageInt = 0
	}

	pagedRecords := allRecords(pageInt)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", pagedRecords)
}

func spotifyRecentlyPlayed(w http.ResponseWriter, r *http.Request) {
	formattedData, _, err := getSpotifyRecentlyPlayed()

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		jsonError(w, r, 500, fmt.Sprintf("%s", err))
	} else {
		fmt.Fprintf(w, "%s\n", string(formattedData))
	}
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

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code, _ := r.URL.Query()["code"]
	state, _ := r.URL.Query()["state"]
	err, _ := r.URL.Query()["error"]

	_ = code
	_ = state
	_ = err

	if len(err) > 0 {
		// Todo
	}

	spotifyAuthUrl := "https://accounts.spotify.com/api/token"

	spotifyAuthPayload := url.Values{}
	spotifyAuthPayload.Set("grant_type", "authorization_code")
	spotifyAuthPayload.Set("code", code[0])
	spotifyAuthPayload.Set("redirect_uri", sCallbackUrl)
	spotifyAuthPayload.Set("client_id", sClientId)
	spotifyAuthPayload.Set("client_secret", sClientSecret)

	httpClient := &http.Client{}

	sr, _ := http.NewRequest(http.MethodPost, spotifyAuthUrl, strings.NewReader(spotifyAuthPayload.Encode()))
	sr.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	spotifyResp, _ := httpClient.Do(sr)

	if spotifyResp.Body != nil {
		defer spotifyResp.Body.Close()
	}

	spotifyRespBody, srbErr := ioutil.ReadAll(spotifyResp.Body)
	checkErr(srbErr)

	spotifyRespBodyParsed := SpotifyAuth{}
	jsonErr := json.Unmarshal(spotifyRespBody, &spotifyRespBodyParsed)
	checkErr(jsonErr)

	fmt.Println(spotifyRespBodyParsed)

	t := time.Now()

	setConf("ACCESS", spotifyRespBodyParsed.AccessToken)
	setConf("REFRESH", spotifyRespBodyParsed.RefreshToken)
	setConf("ACCESS_VALIDITY", strconv.FormatInt(t.Unix()+spotifyRespBodyParsed.ExpiresIn, 10))
}

/* misc */

func ensureToken() error {
	accessValidity, avErr := strconv.ParseInt(getConf("ACCESS_VALIDITY"), 10, 64)
	checkErr(avErr)

	t := time.Now()

	timeDiff := accessValidity - t.Unix()

	if timeDiff < 60 {
		glog.Info("Getting new token.")
		spotifyAuthUrl := "https://accounts.spotify.com/api/token"

		spotifyAuthPayload := url.Values{}
		spotifyAuthPayload.Set("grant_type", "refresh_token")
		spotifyAuthPayload.Set("refresh_token", getConf("REFRESH"))
		spotifyAuthPayload.Set("client_id", sClientId)
		spotifyAuthPayload.Set("client_secret", sClientSecret)

		httpClient := &http.Client{}

		sr, _ := http.NewRequest(
			http.MethodPost,
			spotifyAuthUrl,
			strings.NewReader(spotifyAuthPayload.Encode()),
		)
		sr.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		spotifyResp, spotifyErr := httpClient.Do(sr)

		if spotifyErr != nil {
			return spotifyErr
		}

		if spotifyResp.Body != nil {
			defer spotifyResp.Body.Close()
		}

		spotifyRespBody, srbErr := ioutil.ReadAll(spotifyResp.Body)

		if spotifyResp.StatusCode > 299 {
			glog.V(3).Info(fmt.Sprintf("Error in token resp: %s", spotifyRespBody))
		}

		checkErr(srbErr)

		spotifyRespBodyParsed := SpotifyAuth{}
		jsonErr := json.Unmarshal(spotifyRespBody, &spotifyRespBodyParsed)
		checkErr(jsonErr)

		setConf("ACCESS", spotifyRespBodyParsed.AccessToken)
		setConf("ACCESS_VALIDITY", strconv.FormatInt(t.Unix()+spotifyRespBodyParsed.ExpiresIn, 10))
		glog.Info("New token expires in: ", getConf("ACCESS_VALIDITY"))
	}

	return nil
}

func getSpotifyRecentlyPlayed() (string, SpotifyRecentlyPlayed, error) {
	spotifyApiUrl := "https://api.spotify.com/v1/me/player/recently-played"
	var respStruct SpotifyRecentlyPlayed

	tokenErr := ensureToken()

	if tokenErr != nil {
		return "", respStruct, tokenErr
	}

	bearerHeader := fmt.Sprintf("Bearer %s", getConf("ACCESS"))

	httpClient := &http.Client{}

	sr, _ := http.NewRequest(http.MethodGet, spotifyApiUrl, nil)
	sr.Header.Add("Authorization", bearerHeader)

	spotifyResp, spotifyErr := httpClient.Do(sr)

	if spotifyErr != nil {
		glog.V(3).Info(spotifyErr)
		return "", respStruct, spotifyErr
	}

	data, _ := ioutil.ReadAll(spotifyResp.Body)
	glog.V(3).Info(fmt.Sprintf("Spotify resp: %s", data))

	if spotifyResp.Body != nil {
		defer spotifyResp.Body.Close()
	}

	jsonErr := json.Unmarshal(data, &respStruct)
	checkErr(jsonErr)

	formattedData, formattedDataErr := json.Marshal(respStruct.Items)
	checkErr(formattedDataErr)

	return string(formattedData), respStruct, nil
}

func syncData() {
	_, incomingData, incomingErr := getSpotifyRecentlyPlayed()

	if incomingErr != nil {
		glog.Errorf("%s", incomingErr)
		return
	}

	for i := len(incomingData.Items) - 1; i >= 0; i-- {
		recentTrack := incomingData.Items[i]
		checkQuery, checkQueryErr := musicDB.Prepare("SELECT * FROM music WHERE played_at = ?")
		defer checkQuery.Close()
		checkErr(checkQueryErr)

		var track Track

		checkRowErr := checkQuery.QueryRow(recentTrack.PlayedAt).Scan(&track.Id, &track.Artist, &track.Album, &track.Name, &track.Uri, &track.Added, &track.PlayedAt)
		switch {
		case checkRowErr == sql.ErrNoRows:
			glog.Infof("Inserting %s by %s.\n", recentTrack.Track.Name, recentTrack.Track.Artists[0].Name)
			addRecord(
				recentTrack.Track.Artists[0].Name,
				recentTrack.Track.Album.Name,
				recentTrack.Track.Name,
				recentTrack.Track.URI,
				recentTrack.PlayedAt,
			)
		case checkRowErr != nil:
			checkErr(checkRowErr)
		default:
			glog.V(3).Info(fmt.Sprintf("NOT inserting %s by %s.\n", recentTrack.Track.Name, recentTrack.Track.Artists[0].Name))
		}
	}
}

/* main */

func init() {
	flag.Parse()
}

func main() {
	var allEnvs bool

	sClientId, allEnvs = os.LookupEnv(ENV1)
	sClientSecret, allEnvs = os.LookupEnv(ENV2)
	sCallbackUrl, allEnvs = os.LookupEnv(ENV3)
	sDbPath, allEnvs = os.LookupEnv(ENV4)

	dbUri := fmt.Sprintf("file:%s?_mutex=full", sDbPath)

	if !allEnvs {
		glog.Fatal("Please declare all variables!")
	}

	db, err := sql.Open("sqlite3", dbUri)
	defer db.Close()
	checkErr(err)
	musicDB = db

	statementMusic, errMusic := db.Prepare(
		`CREATE TABLE IF NOT EXISTS music 
		(id INTEGER PRIMARY KEY, artist TEXT, 
		album TEXT, name TEXT, uri TEXT, 
		add_time CURRENT_TIMESTAMP, played_at TEXT)`)
	checkErr(errMusic)
	statementMusic.Exec()

	statementConf, errConf := db.Prepare(
		`CREATE TABLE IF NOT EXISTS conf
		(key TEXT PRIMARY KEY, value TEXT)`)
	checkErr(errConf)
	statementConf.Exec()

	statementConfBlank, errConfBlank := db.Prepare(
		`INSERT OR IGNORE INTO conf (key, value) 
		VALUES ("ACCESS",""), ("REFRESH",""), ("ACCESS_VALIDITY", "")`)
	checkErr(errConfBlank)
	statementConfBlank.Exec()

	mainRouter := mux.NewRouter()
	apiRouter := mainRouter.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/tracks", insert).Methods("POST")
	apiRouter.HandleFunc("/tracks", getAllPaged).Methods("GET")
	apiRouter.HandleFunc("/tracks/{id}", getById).Methods("GET")
	apiRouter.HandleFunc("/s/auth", spotifyAuthorize).Methods("GET")
	apiRouter.HandleFunc("/callback", callbackHandler).Methods("GET")
	apiRouter.HandleFunc("/s/history", spotifyRecentlyPlayed).Methods("GET")

	mainRouter.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("html"))))

	srv := &http.Server{
		Handler:      mainRouter,
		Addr:         "0.0.0.0:6789",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	glog.V(2).Info("Client: ", sClientId)
	glog.V(2).Info("Secret: ", sClientSecret)
	glog.V(2).Info("Callback: ", sCallbackUrl)

	glog.V(2).Info("PID: ", os.Getpid())

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			glog.V(3).Info("Syncing data...")
			syncData()
		}
	}()

	defer ticker.Stop()

	checkErr(srv.ListenAndServe())
}
