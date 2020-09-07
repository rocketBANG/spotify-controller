package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config contains the config for this app
type Config struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// Load will load the current config
func Load() Config {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	return config
}
