package spotify

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Returns false if code should not continue
// true if code should continue
func handleError(err *ErrorResult) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Pause will pause spotify
func Pause() {
	makeAuthReq("PUT", "https://api.spotify.com/v1/me/player/pause", nil)
}

// Play will play spotify
func Play() {
	makeAuthReq("PUT", "https://api.spotify.com/v1/me/player/play", nil)
}

// Next will go to the next track
func Next() {
	makeAuthReq("POST", "https://api.spotify.com/v1/me/player/next", nil)
}

// Prev will go to the previous track
func Prev() {
	makeAuthReq("POST", "https://api.spotify.com/v1/me/player/previous", nil)
}

// GetPlaylists will get the playlists for the current spotify user
func GetPlaylists() []*Playlist {
	playlists := &playlistReq{}
	err := tryMakeReq("GET", "https://api.spotify.com/v1/me/playlists", playlists)
	if !handleError(err) {
		return nil
	}
	convertedPlaylists := make([]*Playlist, len(playlists.Items))
	for i, playlist := range playlists.Items {
		convertedPlaylists[i] = &Playlist{
			ID:   playlist.ID,
			Name: playlist.Name,
		}

		fmt.Println(playlist.Name)
	}
	return convertedPlaylists
}

// GetCurrentSong will get the currently playing song or nil if there is no song playing
func GetCurrentSong() *Song {
	currentlyPlaying := getCurrentlyPlaying()

	if currentlyPlaying.CurrentlyPlayingType != "track" {
		fmt.Println("No track playing")
		return nil
	}

	return &Song{
		Name: currentlyPlaying.Item.Name,
		URI:  currentlyPlaying.Item.URI,
	}
}

// GetCurrentPlaylist returns the ID from the current playlist
func GetCurrentPlaylist() *Playlist {
	currentlyPlaying := getCurrentlyPlaying()

	if currentlyPlaying.Context.Type != "playlist" {
		fmt.Println("Could not get current playlist")
		return nil
	}

	return &Playlist{
		ID: getIDFromURI(currentlyPlaying.Context.URI),
	}
}

// AddToPlaylist will attempt to add the given song to the playlist
func AddToPlaylist(playlistID string, songURI string) {
	// TODO need to check if item already exists

	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?uris=%s", playlistID, songURI)
	err := tryMakeReq("POST", url, nil)
	if !handleError(err) {
		return
	}

}

// RemoveFromPlaylist will attempt to remove the given song from the given playlist
func RemoveFromPlaylist(playlistID string, songURI string) {
	body := map[string][]deletePlaylistBody{
		"tracks": []deletePlaylistBody{deletePlaylistBody{
			URI: songURI,
		}},
	}

	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?uris=%s", playlistID, songURI)
	err := tryMakeReq2("DELETE", url, nil, body)
	if !handleError(err) {
		return
	}

}

// SetVolume will change the spotify volume to the given value
func SetVolume(percent int) {
	percentStr := strconv.Itoa(percent)
	makeAuthReq("PUT", "https://api.spotify.com/v1/me/player/volume?volume_percent="+percentStr, nil)
}

func getIDFromURI(URI string) string {
	splitURI := strings.Split(URI, ":")
	if len(splitURI) == 0 {
		log.Fatalf("Could not find id from URI %s\n", splitURI)
		return ""
	}
	return splitURI[len(splitURI)-1]
}

func getCurrentlyPlaying() *currentlyPlayingRes {
	currentlyPlaying := &currentlyPlayingRes{}
	err := tryMakeReq("GET", "https://api.spotify.com/v1/me/player/currently-playing", currentlyPlaying)
	if !handleError(err) {
		return nil
	}
	return currentlyPlaying
}

type playlistReq struct {
	Href  string `json:"href"`
	Items []struct {
		Collaborative bool `json:"collaborative"`
		ExternalUrls  struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string        `json:"href"`
		ID     string        `json:"id"`
		Images []interface{} `json:"images"`
		Name   string        `json:"name"`
		Owner  struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"owner"`
		Public     bool   `json:"public"`
		SnapshotID string `json:"snapshot_id"`
		Tracks     struct {
			Href  string `json:"href"`
			Total int    `json:"total"`
		} `json:"tracks"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"items"`
	Limit    int         `json:"limit"`
	Next     interface{} `json:"next"`
	Offset   int         `json:"offset"`
	Previous interface{} `json:"previous"`
	Total    int         `json:"total"`
}

// Song represents a spotify song
type Song struct {
	Name string
	URI  string
}

// Playlist represents a spotify playlist
type Playlist struct {
	Name string
	ID   string
}

type deletePlaylistBody struct {
	URI string `json:"uri"`
}

type currentlyPlayingRes struct {
	Context struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"context"`
	Timestamp            int64  `json:"timestamp"`
	ProgressMs           int    `json:"progress_ms"`
	IsPlaying            bool   `json:"is_playing"`
	CurrentlyPlayingType string `json:"currently_playing_type"`
	Actions              struct {
		Disallows struct {
			Resuming bool `json:"resuming"`
		} `json:"disallows"`
	} `json:"actions"`
	Item struct {
		Album struct {
			AlbumType    string `json:"album_type"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"album"`
		Artists []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		DiscNumber       int      `json:"disc_number"`
		DurationMs       int      `json:"duration_ms"`
		Explicit         bool     `json:"explicit"`
		ExternalIds      struct {
			Isrc string `json:"isrc"`
		} `json:"external_ids"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		Popularity  int    `json:"popularity"`
		PreviewURL  string `json:"preview_url"`
		TrackNumber int    `json:"track_number"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
	} `json:"item"`
}
