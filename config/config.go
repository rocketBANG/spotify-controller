package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config contains the config for this app
type Config struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Debug        bool   `json:"debug"`
}

// Load will load the current config
func Load() error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	dirAbsPath := filepath.Dir(ex)

	file, fileErr := os.Open("config.dev.json")
	if os.IsNotExist(fileErr) {
		file, fileErr = os.Open(filepath.Join(dirAbsPath, "config.json"))
	}

	if os.IsNotExist(fileErr) {
		return errors.New("No config.json found, refer to readme to setup config.json")
	}

	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
		return errors.New("Could not decode config.json")
	}
	Value = config
	return nil
}

// Value is the current config value
var Value Config
