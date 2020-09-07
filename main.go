package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"github.com/rocketbang/spotify-controller/command"
	"github.com/rocketbang/spotify-controller/spotify"
)

var srv *http.Server

func callback(w http.ResponseWriter, req *http.Request) {
	codes, ok := req.URL.Query()["code"]
	if !ok || len(codes) < 1 {
		log.Println("No code returned")
		return
	}

	code := codes[0]
	fmt.Printf("code %s\n", code)

	spotify.Authorise(code)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	fmt.Fprintf(w, "<script>window.close()</script>")
}

func createScopes(scopes []string) string {
	scopeString := ""
	for i, scope := range scopes {
		if i == 0 {
			continue
		}
		scopeString = scopeString + "%20" + scope
	}
	return scopeString
}

func main() {
	fmt.Println("Go")
	scopeString := createScopes([]string{
		"user-read-private",
		"user-read-email",
		"user-read-playback-state",
		"user-modify-playback-state",
		"user-read-currently-playing",
		"streaming",
		"app-remote-control",
		"playlist-read-collaborative",
		"playlist-modify-public",
		"playlist-read-private",
		"playlist-modify-private",
		"user-library-modify",
		"user-library-read",
		"user-top-read",
		"user-read-playback-position",
		"user-read-recently-played",
		"user-follow-read",
		"user-follow-modify",
	})
	browser.OpenURL("https://accounts.spotify.com/authorize?client_id=401772871f2b4065822277a15d71e6d2&response_type=code&redirect_uri=http%3A%2F%2Flocalhost%3A8282%2Fcallback&scope=" + scopeString)

	http.HandleFunc("/callback", callback)

	srv = &http.Server{Addr: ":8282"}
	srv.ListenAndServe()

	command.Listen()

}
