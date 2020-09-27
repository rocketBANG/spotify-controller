package command

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/rocketbang/spotify-controller/spotify"
)

func volume(command string) {
	var volInt int
	var err error
	if len(command) < len("volume")+2 {
		fmt.Printf("Enter volume level (0-100):\n")
		volInt, err = getInt(0, 100)
		if err != nil {
			return
		}
	} else {
		volStr := command[len("volume")+1:]
		volInt, err = strconv.Atoi(volStr)
		if err != nil {
			fmt.Println("Could not read volume level")
			return
		}
	}

	if volInt < 0 || volInt > 100 {
		fmt.Println("Volume level is out of bounds. Use a number between 0 and 100")
		return
	}
	fmt.Printf("Setting volume to %d\n", volInt)
	spotify.SetVolume(volInt)
}

func addToPlaylist() {
	playlist := choosePlaylist()
	if playlist == nil {
		return
	}

	song := spotify.GetCurrentSong()
	if song == nil {
		fmt.Println("Could not get current song")
		return
	}

	fmt.Printf("Adding %s to %s\n", song.Name, playlist.Name)

	spotify.AddToPlaylist(playlist.ID, song.URI)
}

func removeFromCurrentPlaylist() {
	playlist := spotify.GetCurrentPlaylist()
	if playlist == nil {
		fmt.Println("Could not get current playlist")
		return
	}

	song := spotify.GetCurrentSong()
	if song == nil {
		fmt.Println("Could not get current song")
		return
	}

	fmt.Printf("Removing %s from current playlist\n", song.Name)

	spotify.RemoveFromPlaylist(playlist.ID, song.URI, nil)
}

func removeDuplicatesInPlaylist() {
	playlist := choosePlaylist()
	if playlist == nil {
		return
	}

	items := spotify.GetTracksInPlaylist(playlist.ID)

	trackMap := make(map[string]*spotify.PlaylistTrackResItem)

	detectedDuplicate := false
	for i := range items {
		item := items[i]
		track := item.Track
		prevItem := trackMap[track.ID]

		if trackMap[track.ID] != nil {
			detectedDuplicate = true
			fmt.Printf("Found duplicate! %s, %s\n", track.Name, item.AddedAt)
			fmt.Printf("Previous item %s, %s\n", prevItem.Track.Name, prevItem.AddedAt)
			fmt.Printf("Remove duplicate? (y/n)\n")
			remove := getConfirm()
			if remove {
				spotify.RemoveFromPlaylist(playlist.ID, track.URI, []int{i})
			}
		} else {
			trackMap[track.ID] = item
		}
	}

	if !detectedDuplicate {
		fmt.Printf("No duplicates found in %s\n", playlist.Name)
	}

}

func shuffleInNewPlaylist() {
	fmt.Println("Choose playlist to shuffle")
	clonedPlaylist := choosePlaylist()

	items := spotify.GetTracksInPlaylist(clonedPlaylist.ID)
	itemURIs := make([]string, len(items))
	for i, item := range items {
		itemURIs[i] = item.Track.URI
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(itemURIs), func(i, j int) { itemURIs[i], itemURIs[j] = itemURIs[j], itemURIs[i] })

	max := 800
	if max > len(itemURIs)-1 {
		max = len(itemURIs) - 1
	}
	spotify.PlayTracks(itemURIs[0:max], clonedPlaylist.URI)
}

func clonePlaylist() {
	userID, err := spotify.GetUserID()
	if err != nil {
		fmt.Println("Could not get user id")
		return
	}

	fmt.Println("Choose playlist to clone")
	clonedPlaylist := choosePlaylist()

	fmt.Println("Enter new playlist name (New Playlist)")
	playlistName := getString("New Playlist")

	playlistID, err := spotify.CreateNewPlaylist(userID, playlistName, false)
	if err != nil {
		fmt.Println("Could not create new playlist")
		return
	}

	items := spotify.GetTracksInPlaylist(clonedPlaylist.ID)
	itemURIs := make([]string, len(items))
	for i, item := range items {
		itemURIs[i] = item.Track.URI
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(itemURIs), func(i, j int) { itemURIs[i], itemURIs[j] = itemURIs[j], itemURIs[i] })

	spotify.AddManyToPlaylist(playlistID, itemURIs)
}

func help() {
	fmt.Println("Possible commands are:")
	fmt.Printf("pause, play, list, add, remove, exit, shuffle, clone, next, prev, duplicate, track\n")
}

func choosePlaylist() *spotify.Playlist {
	playlists := spotify.GetPlaylists()

	if playlists == nil {
		fmt.Println("Could not get playlists")
		return nil
	}

	fmt.Printf("Choose Playlist:\n")
	for i, playlist := range playlists {
		fmt.Printf("%d. %s\n", (i + 1), playlist.Name)
	}

	playlistNum, err := getInt(1, len(playlists))

	if err != nil {
		return nil
	}

	playlist := playlists[playlistNum-1]
	if playlist == nil {
		fmt.Println("Could not get playlist")
		return nil
	}
	return playlist
}

func getInt(min int, max int) (int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	intStr := scanner.Text()

	number, err := strconv.Atoi(intStr)
	if err != nil {
		fmt.Println("Could not read number")
		return 0, errors.New("")
	}

	if number < min || number > max {
		fmt.Println("Number is out of bounds")
		return 0, errors.New("")
	}

	return number, nil
}

func getString(auto string) string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	str := scanner.Text()

	if str == "" {
		return auto
	}

	return str
}

func getConfirm() bool {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	str := scanner.Text()

	if str == "y" || str == "Y" || str == "yes" {
		return true
	}
	return false
}

func Listen() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "pause" {
			fmt.Println("Pausing music")
			spotify.Pause()
		} else if scanner.Text() == "play" {
			fmt.Println("Playing music")
			spotify.Play()
		} else if scanner.Text() == "list" {
			fmt.Println("Listing playlists music")
			spotify.GetPlaylists()
		} else if startsWith(scanner.Text(), "volume ") || scanner.Text() == "vol" {
			volume(scanner.Text())
		} else if scanner.Text() == "add" {
			addToPlaylist()
		} else if scanner.Text() == "remove" {
			removeFromCurrentPlaylist()
		} else if scanner.Text() == "exit" {
			fmt.Println("Exiting...")
			return
		} else if scanner.Text() == "shuffle" {
			fmt.Println("Creating new playlist to shuffle")
			shuffleInNewPlaylist()
		} else if scanner.Text() == "clone" {
			clonePlaylist()
		} else if scanner.Text() == "help" {
			help()
		} else if scanner.Text() == "next" {
			fmt.Println("Next track")
			spotify.Next()
		} else if scanner.Text() == "prev" {
			fmt.Println("Previous track")
			spotify.Prev()
		} else if scanner.Text() == "duplicate" {
			fmt.Println("Detecting duplicates")
			removeDuplicatesInPlaylist()
		} else if scanner.Text() == "track" {
			song := spotify.GetCurrentSong()
			fmt.Printf("Current track is: %s by %s. - %s\n", song.Name, song.PrimaryArtist, song.Album)
		} else {
			fmt.Println("Unrecongised command")
		}

		// fmt.Println(scanner.Text())
	}

	if scanner.Err() != nil {
		log.Fatal("Scanner error")
	}

}

func startsWith(command string, search string) bool {
	if len(command) < len(search) {
		return false
	}
	return command[0:len(search)] == search
}
