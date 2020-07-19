package main

type SpotifyRecentlyPlayed struct {
	Cursors struct {
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"cursors"`
	Href  string `json:"href"`
	Items []struct {
		Context struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"-"`
			Href string `json:"href"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"-"`
		PlayedAt string `json:"played_at"`
		Track    struct {
			Album struct {
				AlbumType string `json:"album_type"`
				Artists   []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"-"`
					Href string `json:"-"`
					ID   string `json:"-"`
					Name string `json:"name"`
					Type string `json:"-"`
					URI  string `json:"-"`
				} `json:"artists"`
				AvailableMarkets []string `json:"-"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"-"`
				Href   string `json:"-"`
				ID     string `json:"id"`
				Images []struct {
					Height int64  `json:"height"`
					URL    string `json:"url"`
					Width  int64  `json:"width"`
				} `json:"images"`
				Name                 string `json:"name"`
				ReleaseDate          string `json:"release_date"`
				ReleaseDatePrecision string `json:"-"`
				TotalTracks          int64  `json:"-"`
				Type                 string `json:"-"`
				URI                  string `json:"-"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"-"`
				Href string `json:"-"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"-"`
				URI  string `json:"-"`
			} `json:"artists"`
			AvailableMarkets []string `json:"-"`
			DiscNumber       int64    `json:"-"`
			DurationMs       int64    `json:"-"`
			Explicit         bool     `json:"-"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"-"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"-"`
			Href        string `json:"-"`
			ID          string `json:"id"`
			IsLocal     bool   `json:"-"`
			Name        string `json:"name"`
			Popularity  int64  `json:"-"`
			PreviewURL  string `json:"-"`
			TrackNumber int64  `json:"-"`
			Type        string `json:"-"`
			URI         string `json:"uri"`
		} `json:"track"`
	} `json:"items"`
	Limit int64  `json:"limit"`
	Next  string `json:"next"`
}

type Track struct {
	Id       int64  `json:"id"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Name     string `json:"name"`
	Uri      string `json:"uri"`
	Added    string `json:"add_time"`
	PlayedAt string `json:"played_at"`
}

type SpotifyAuth struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
