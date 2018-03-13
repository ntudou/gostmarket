package main

import (
	"encoding/json"
	"os"
)

// Config for client
type Config struct {
	Listen string `json:"listen"`
	Key    string `json:"key"`
}

func parseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}
