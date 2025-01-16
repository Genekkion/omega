package structs

import (
	"encoding/json"
	"os"
)

type Config struct {
	Ignore   []string `json:"ignore"`
	Commands []string `json:"commands"`
	LogLevel string   `json:"logLevel"`
	LogFiles []string `json:"logFiles"`
	Timeout  int      `json:"timeout"`
	Delay    int      `json:"delay"`
}

func ConfigFromFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c Config) WriteToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
