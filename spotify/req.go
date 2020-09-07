package spotify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rocketbang/spotify-controller/config"
)

var authRes *AuthResult

func makeAuthReq(method string, url string) *http.Response {
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)

	// TODO more error handling

	req.Header.Add("Authorization", "Bearer "+authRes.AccessToken)
	res, _ := client.Do(req)

	if res.StatusCode == 401 {
		log.Println("Refreshing...")
		refresh(authRes.RefreshToken)

		req2, _ := http.NewRequest(method, url, nil)
		req2.Header.Add("Authorization", "Bearer "+authRes.AccessToken)
		res, _ = client.Do(req2)
	}

	return res
}

func tryMakeReq(method string, url string, result interface{}) *ErrorResult {
	resp := makeAuthReq(method, url)

	if resp.StatusCode == 400 {
		errorResult := &ErrorResult{}
		json.NewDecoder(resp.Body).Decode(errorResult)
		fmt.Println(errorResult)
		return errorResult
	}

	if result != nil {
		json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func refresh(refreshToken string) {
	config := config.Load()

	client := &http.Client{}

	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("redirect_uri", "http://localhost:8282/callback")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	if err != nil {
		log.Fatal("Could not create refresh token request")
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal("Could not send refresh token request")
		return
	}

	if resp.StatusCode == 400 {
		errorResult := &ErrorResult{}
		json.NewDecoder(resp.Body).Decode(errorResult)
		fmt.Println(errorResult)
		return
	}

	result := &AuthResult{}
	json.NewDecoder(resp.Body).Decode(result)
	if result.RefreshToken == "" {
		result.RefreshToken = authRes.RefreshToken
	}
	authRes = result
}

func Authorise(code string) {
	config := config.Load()

	client := &http.Client{}

	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "http://localhost:8282/callback")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("error when sending req")
	}

	result := &AuthResult{}
	fmt.Println(resp.Status)
	fmt.Println(resp.StatusCode)

	if resp.StatusCode == 400 {
		errorResult := &ErrorResult{}
		json.NewDecoder(resp.Body).Decode(errorResult)
		fmt.Println(errorResult)
		return
	}

	json.NewDecoder(resp.Body).Decode(result)
	fmt.Println(result)
	authRes = result

}

type AuthResult struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// AuthResult is the result from the spotify auth method
type ErrorResult struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type ReqError struct {
	Error ReqErrorDetails `json:"error"`
}

type ReqErrorDetails struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
