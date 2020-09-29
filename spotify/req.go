package spotify

import (
	"bytes"
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

func makeSimpleReq(method string, url string, bodyBuffer *bytes.Buffer) (*http.Request, error) {
	if bodyBuffer == nil {
		return http.NewRequest(method, url, nil)
	}

	return http.NewRequest(method, url, bodyBuffer)
}

func makeAuthReq(method string, url string, bodyBuffer *bytes.Buffer) *http.Response {
	client := &http.Client{}
	req, _ := makeSimpleReq(method, url, bodyBuffer)

	// TODO more error handling

	req.Header.Add("Authorization", "Bearer "+authRes.AccessToken)
	res, _ := client.Do(req)

	if res.StatusCode == 401 {
		log.Println("Refreshing...")
		refresh(authRes.RefreshToken)

		req2, _ := makeSimpleReq(method, url, bodyBuffer)
		req2.Header.Add("Authorization", "Bearer "+authRes.AccessToken)
		res, _ = client.Do(req2)
	}

	return res
}

func tryMakeReq(method string, url string, result interface{}) *ErrorResult {
	return tryMakeReq2(method, url, result, nil)
}

func tryMakeReq2(method string, url string, result interface{}, body interface{}) *ErrorResult {
	var bodyBytes *bytes.Buffer = nil
	if body != nil {
		body, err := json.Marshal(body)
		if err != nil {
			log.Fatal("Could not Marshal body req")
			return nil
		}
		bodyBytes = bytes.NewBuffer(body)
	}

	resp := makeAuthReq(method, url, bodyBytes)

	if resp.StatusCode == 400 {
		errorResult := &ErrorResult{}
		json.NewDecoder(resp.Body).Decode(errorResult)
		fmt.Println(errorResult)
		return errorResult
	}

	defer resp.Body.Close()

	if result != nil {
		json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func refresh(refreshToken string) {
	config := config.Value

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

// Authorise will bring up an authorisation dialog for the current user and then populate authRes
func Authorise(code string) {
	config := config.Value

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

	if config.Debug {
		fmt.Println(resp.Status)
		fmt.Println(resp.StatusCode)
	}

	if resp.StatusCode == 400 {
		errorResult := &ErrorResult{}
		json.NewDecoder(resp.Body).Decode(errorResult)
		fmt.Println(errorResult)
		return
	}

	json.NewDecoder(resp.Body).Decode(result)

	if config.Debug {
		fmt.Println(result)
	}

	authRes = result
}

// AuthResult is the result from the spotify auth method
type AuthResult struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// ErrorResult is the generic result type from a auth route
type ErrorResult struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ReqError is the basic error from a route
type ReqError struct {
	Error ReqErrorDetails `json:"error"`
}

// ReqErrorDetails are the specifics of an error from a route
type ReqErrorDetails struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
