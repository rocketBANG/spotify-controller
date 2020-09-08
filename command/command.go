package command

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

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
	playlists := spotify.GetPlaylists()

	if playlists == nil {
		fmt.Println("Could not get playlists")
		return
	}

	fmt.Printf("Choose Playlist:\n")
	for i, playlist := range playlists {
		fmt.Printf("%d. %s\n", (i + 1), playlist.Name)
	}

	playlistNum, err := getInt(1, len(playlists))

	if err != nil {
		return
	}

	playlist := playlists[playlistNum-1]
	if playlist == nil {
		fmt.Println("Could not get playlist")
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

	spotify.RemoveFromPlaylist(playlist.ID, song.URI)
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
		} else if scanner.Text() == "next" {
			fmt.Println("Next track")
			spotify.Next()
		} else if scanner.Text() == "prev" {
			fmt.Println("Previous track")
			spotify.Prev()
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
