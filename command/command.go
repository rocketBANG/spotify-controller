package command

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rocketbang/spotify-controller/spotify"
)

func volume(args string) {
	var volInt int
	var err error
	if args == "" {
		fmt.Printf("Enter volume level (0-100):\n")
		volInt, err = getInt(0, 100)
		if err != nil {
			return
		}
	} else {
		volInt, err = strconv.Atoi(args)
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
	if clonedPlaylist == nil {
		return
	}

	items := spotify.GetTracksInPlaylist(clonedPlaylist.ID)
	itemURIs := make([]string, len(items))
	for i, item := range items {
		itemURIs[i] = item.Track.URI
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(itemURIs), func(i, j int) { itemURIs[i], itemURIs[j] = itemURIs[j], itemURIs[i] })

	spotify.PlayTracks(itemURIs, clonedPlaylist.URI)
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

func playingStatus() {
	song := spotify.GetCurrentSong()
	fmt.Printf("Current track is: %s by %s. - %s\n", song.Name, song.PrimaryArtist, song.Album)
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

func getArgs(text string, cmdText string) string {
	if text == cmdText {
		return ""
	}
	if startsWith(text, cmdText+" ") {
		return text[len(cmdText)+1:]
	}
	return ""
}

func findMatchedCommand(commands []*commandStruct, text string) (*commandStruct, string) {
	for _, command := range commands {
		for _, cmdText := range command.CmdText {
			if text == cmdText || startsWith(text, cmdText+" ") {
				return command, cmdText
			}
		}
	}
	return nil, ""
}

func runCommand(commands []*commandStruct, text string) bool {
	command, cmdText := findMatchedCommand(commands, text)
	if command == nil {
		return false
	}

	if command.RunText != "" {
		fmt.Println(command.RunText)
	}
	args := getArgs(text, cmdText)
	command.Run(args)
	return true
}

func printHelpCommand(command *commandStruct) {
	fmt.Println("\n" + command.Help + "\n")
}

func printHelp(commands []*commandStruct, args string) {
	if args != "" {
		command, _ := findMatchedCommand(commands, args)
		if command == nil {
			fmt.Println("Could not find the given command")
			return
		}
		printHelpCommand(command)
		return
	}
	fmt.Println("\nAvailable commands are: ")
	for _, command := range commands {
		firstLine := strings.Split(command.Help, "\n")[0]
		fmt.Println(command.CmdText[0] + " - " + firstLine)
	}
	fmt.Print("\nFor more information use 'help [command]'\n")
}

// Listen will listen for the given commands until the user exits
func Listen() {
	commands := make([]*commandStruct, 0)

	commands = append(commands, &commandStruct{
		Name:    "Pause",
		Help:    "Use to pause the music",
		Run:     func(a string) { spotify.Pause() },
		CmdText: []string{"pause"},
		RunText: "Pausing Music",
	})
	commands = append(commands, &commandStruct{
		Name:    "Play",
		Help:    "Use to play the music",
		Run:     func(a string) { spotify.Play() },
		CmdText: []string{"play"},
		RunText: "Playing Music",
	})
	commands = append(commands, &commandStruct{
		Name:    "Volume",
		Help:    "Use to raise or lower the volume\nCan use either volume or vol",
		Run:     volume,
		CmdText: []string{"volume", "vol"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Remove",
		Help:    "Removes the currently playing song from the current playlist",
		Run:     func(a string) { removeFromCurrentPlaylist() },
		CmdText: []string{"remove"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Add",
		Help:    "Adds the currently playing song to a playlist of your choice",
		Run:     func(a string) { addToPlaylist() },
		CmdText: []string{"add"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Shuffle",
		Help:    "Randomly shuffles the given playlist",
		Run:     func(a string) { shuffleInNewPlaylist() },
		CmdText: []string{"shuffle"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Clone",
		Help:    "Clones the given playlist to a new playlist with a randomly shuffled order",
		Run:     func(a string) { clonePlaylist() },
		CmdText: []string{"clone"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Next",
		Help:    "Skips to the next track",
		Run:     func(a string) { spotify.Next() },
		RunText: "Next track",
		CmdText: []string{"next"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Previous",
		Help:    "Skips back to the previous track",
		Run:     func(a string) { spotify.Prev() },
		RunText: "Previous track",
		CmdText: []string{"prev", "previous"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Duplicate detect",
		Help:    "Removes duplicates from the given playlist",
		Run:     func(a string) { removeDuplicatesInPlaylist() },
		RunText: "Detecting duplicates",
		CmdText: []string{"duplicate"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Status",
		Help:    "Gets the details of the currently playing track",
		Run:     func(a string) { playingStatus() },
		CmdText: []string{"details", "status", "playing"},
	})
	commands = append(commands, &commandStruct{
		Name:    "Version",
		Help:    "Prints the current version",
		Run:     func(a string) { fmt.Println("0.1.0") },
		CmdText: []string{"version"},
	})

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		didRun := runCommand(commands, scanner.Text())
		if didRun {
			continue
		}

		if scanner.Text() == "exit" {
			fmt.Println("Exiting...")
			return
		} else if scanner.Text() == "" {
			continue
		} else if startsWith(scanner.Text(), "help") {
			printHelp(commands, getArgs(scanner.Text(), "help"))
		} else {
			fmt.Println("Unrecongised command")
		}
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

type commandStruct struct {
	Name    string
	Help    string
	CmdText []string
	Run     func(string)
	RunText string
}
