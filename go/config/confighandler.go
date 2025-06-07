package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type Resolution struct {
	Width  int
	Height int
}

type Config struct {
	Resolution 		Resolution 	`json:"resolution"`
	Profile    		string     	`json:"profile"`
	DefaultWebpage 	string 		`json:"defaultWebpage"`
}

func LoadConfig() *Config {
	defaultConfig := Config{
		Resolution: Resolution{530, 900},
		Profile:    "Profile",
		DefaultWebpage: "https://google.com",
	}

	configFile, err := os.ReadFile("config.json")

	if err != nil {
		return &defaultConfig
	}

	json.Unmarshal(configFile, &defaultConfig)

	return &defaultConfig
}

func (r *Resolution) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}

	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return nil
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil
	}

	r.Width = width
	r.Height = height

	return nil
}
