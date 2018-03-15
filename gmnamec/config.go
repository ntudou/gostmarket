package main

import (
	"encoding/json"
	"os"
)

// Config for client
type Config struct {
	Listen    string `json:"listen"`
	Target    string `json:"target"`
	Namespace string `json:"namespace"`
	Interface string `json:interface`
	Port      string `json:"port"`
	Key       string `json:"key"`
	Code      string `json:"code"`
}

func parseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}
