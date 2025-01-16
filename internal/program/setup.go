package program

import (
	"errors"
	"omega/internal/config"
	"omega/internal/env"
	"omega/internal/forms"
	"omega/internal/log"
	"os"
)

func Setup() (*config.Config, error) {
	exists, err := checkWritableFile(env.ConfigPath)
	if err != nil {
		return nil, err
	}

	var c *config.Config

	if exists {
		// Config file supposedly exists, so we check for its validity
		c, err = config.FromFile(env.ConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Config file does not exist, so we see if the user wants one
		if !forms.GenerateConfig() {
			log.Info("Alright see ya later 🐊!")
			os.Exit(0)
		}

		c = &config.DefaultConfig
		err = c.WriteToFile(env.ConfigPath)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Will cowardly assume that a directory is generally the incorrect target path
// and error accordingly
func checkWritableFile(path string) (exists bool, err error) {
	info, err := os.Stat(path)
	// Check if it is because file not found
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File does not exist
			return false, nil
		} else {
			// Some kind of os error perhaps
			return false, err
		}
	}

	// At this point err is nil, means existing file

	// Check if existing file is a dir
	if info.IsDir() {
		return true, errors.New("unable to write file to existing directory")
	}

	return true, nil
}
