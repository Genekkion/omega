package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Ignore   []string `json:"ignore"`
	Commands []string `json:"commands"`
	LogLevel string   `json:"logLevel,omitempty"`
	LogFiles []string `json:"logFiles,omitempty"`
	Timeout  int      `json:"timeout"`
	Delay    int      `json:"delay"`
}

func From(r io.Reader) (*Config, error) {
	var config Config
	err := json.NewDecoder(r).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func FromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return From(file)
}

func (c Config) Write(w io.Writer) error {
	data, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// Assumes writeable file, will just err otherwise
func (c Config) WriteToFile(path string) error {
	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write to file
	return c.Write(file)
}
