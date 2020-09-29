package spotify

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rocketbang/spotify-controller/config"
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

// PlayTracks will play the given tracks
// Has a maximum limit of 800 (any extra will not be included)
func PlayTracks(songURIs []string, playlistURI string) {
	// This maximum isn't documented but the PlayTracks request will constantly fail if max is ~900 so I've set it to 800
	max := 800
	if max > len(songURIs)-1 {
		max = len(songURIs) - 1
	}

	body := &playReq{
		URIs: songURIs[0:max],
	}

	err := tryMakeReq2("PUT", "https://api.spotify.com/v1/me/player/play", nil, body)
	if !handleError(err) {
		return
	}

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
			URI:  playlist.URI,
		}
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

	artist := ""
	if currentlyPlaying.Item.Artists != nil && len(currentlyPlaying.Item.Artists) > 0 {
		artist = currentlyPlaying.Item.Artists[0].Name
	}

	album := currentlyPlaying.Item.Album.Name

	return &Song{
		Name:          currentlyPlaying.Item.Name,
		URI:           currentlyPlaying.Item.URI,
		PrimaryArtist: artist,
		Album:         album,
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

// AddManyToPlaylist will attempt to add the given songs to the playlist with no limit
func AddManyToPlaylist(playlistID string, songURIs []string) {
	offset := 0
	for len(songURIs)-offset > 0 {
		max := offset + 100
		if max > len(songURIs)-1 {
			max = len(songURIs) - 1
		}
		AddManyToPlaylistWithLimit(playlistID, songURIs[offset:max])
		offset = offset + 100
	}
}

// AddManyToPlaylistWithLimit will attempt to add the given songs to the playlist
//
// Limit 100
func AddManyToPlaylistWithLimit(playlistID string, songURIs []string) {
	body := map[string][]string{
		"uris": songURIs,
	}

	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)
	err := tryMakeReq2("POST", url, nil, body)
	if !handleError(err) {
		return
	}
}

// RemoveFromPlaylist will attempt to remove the given song from the given playlist
func RemoveFromPlaylist(playlistID string, songURI string, positions []int) {
	body := map[string][]deletePlaylistBody{
		"name": []deletePlaylistBody{deletePlaylistBody{
			URI:       songURI,
			Positions: positions,
		}},
	}

	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?uris=%s", playlistID, songURI)
	err := tryMakeReq2("DELETE", url, nil, body)
	if !handleError(err) {
		return
	}

}

// CreateNewPlaylist will create the given playlist
func CreateNewPlaylist(userID, playlistName string, isPublic bool) (string, error) {
	body := &createPlaylistBody{
		Name:   playlistName,
		Public: isPublic,
	}

	res := &newPlaylistReq{}

	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userID)
	err := tryMakeReq2("POST", url, res, body)
	if !handleError(err) {
		return "", errors.New("Could not create playlist")
	}

	return res.ID, nil
}

// GetUserID gets the user ID for the currently logged in user
func GetUserID() (string, error) {
	userDetails := getUserDetails()
	if userDetails == nil {
		return "", errors.New("Could not get user ID")
	}
	return userDetails.ID, nil
}

func getUserDetails() *userDetailReq {
	userDetails := &userDetailReq{}
	err := tryMakeReq("GET", "https://api.spotify.com/v1/me", userDetails)
	if !handleError(err) {
		return nil
	}

	return userDetails
}

// SetVolume will change the spotify volume to the given value
func SetVolume(percent int) {
	percentStr := strconv.Itoa(percent)
	makeAuthReq("PUT", "https://api.spotify.com/v1/me/player/volume?volume_percent="+percentStr, nil)
}

// GetTracksInPlaylist gets all the tracks in a playlist
func GetTracksInPlaylist(playlistID string) []*PlaylistTrackResItem {
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?limit=100", playlistID)
	res := &playlistTrackRes{}

	if config.Value.Debug {
		fmt.Printf("fetching: %s\n", url)
	}

	err := tryMakeReq("GET", url, res)
	if !handleError(err) {
		return nil
	}

	items := getPagesAsync(url, res, &playlistTrackRes{})

	return items
}

func getPagesAsync(url string, paging *playlistTrackRes, res interface{}) []*PlaylistTrackResItem {
	var wg sync.WaitGroup

	pageTotal := paging.Total/paging.Limit + 1
	resTotal := make([]*PlaylistTrackResItem, paging.Total)

	i := 0
	for i < pageTotal {
		offset := i * paging.Limit
		wg.Add(1)

		go func() {
			defer wg.Done()
			res := getSinglePage(url, offset)
			for resIndex := range res.Items {
				resTotal[resIndex+offset] = &res.Items[resIndex]
			}
		}()
		i++
	}
	wg.Wait()
	return resTotal
}

func getSinglePage(url string, offset int) *playlistTrackRes {
	parsedURL := fmt.Sprintf("%s&offset=%d", url, offset)
	res := &playlistTrackRes{}

	if config.Value.Debug {
		fmt.Printf("fetching: %s\n", parsedURL)
	}

	err := tryMakeReq("GET", parsedURL, res)
	if !handleError(err) {
		return nil
	}
	return res
}

func getNextPage(paging *playlistTrackRes, res interface{}) interface{} {
	if paging.Next == "" || paging.Total-paging.Limit < paging.Offset {
		return nil
	}

	if config.Value.Debug {
		fmt.Printf("fetching: %s\n", paging.Next)
	}

	err := tryMakeReq("GET", paging.Next, res)
	if !handleError(err) {
		return nil
	}

	// fmt.Printf("limit: %d, offset: %d, total: %d\n", paging.Limit, paging.Offset, paging.Total)
	// fmt.Printf("next: %s\n", paging.Next)

	return res
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
	Name          string
	URI           string
	PrimaryArtist string
	Album         string
}

// Playlist represents a spotify playlist
type Playlist struct {
	Name string
	ID   string
	URI  string
}

type deletePlaylistBody struct {
	URI       string `json:"uri"`
	Positions []int  `json:"positions"`
}

type createPlaylistBody struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
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

// Paging is a result for pageable data
type Paging struct {
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
}

type playlistTrackRes struct {
	Paging
	Href  string                 `json:"href"`
	Items []PlaylistTrackResItem `json:"items"`
}

// PlaylistTrackResItem represents an item in a playlist
type PlaylistTrackResItem struct {
	AddedAt time.Time `json:"added_at"`
	AddedBy struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"added_by"`
	IsLocal bool `json:"is_local"`
	Track   struct {
		Album struct {
			AlbumType string `json:"album_type"`
			Artists   []struct {
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
			ExternalUrls     struct {
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
	} `json:"track,omitempty"`
}

type userDetailReq struct {
	Country      string `json:"country"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href  interface{} `json:"href"`
		Total int         `json:"total"`
	} `json:"followers"`
	Href   string `json:"href"`
	ID     string `json:"id"`
	Images []struct {
		Height interface{} `json:"height"`
		URL    string      `json:"url"`
		Width  interface{} `json:"width"`
	} `json:"images"`
	Product string `json:"product"`
	Type    string `json:"type"`
	URI     string `json:"uri"`
}

type newPlaylistReq struct {
	Collaborative bool        `json:"collaborative"`
	Description   interface{} `json:"description"`
	ExternalUrls  struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href  interface{} `json:"href"`
		Total int         `json:"total"`
	} `json:"followers"`
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
		Href     string        `json:"href"`
		Items    []interface{} `json:"items"`
		Limit    int           `json:"limit"`
		Next     interface{}   `json:"next"`
		Offset   int           `json:"offset"`
		Previous interface{}   `json:"previous"`
		Total    int           `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type playReq struct {
	URIs []string `json:"uris"`
}
