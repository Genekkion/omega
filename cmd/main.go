package main

import (
	"context"
	"errors"
	"io/fs"
	"omega/internal/env"
	"omega/internal/forms"
	"omega/internal/log"
	"omega/internal/program"
	"omega/internal/runner"
	"omega/internal/structs"
	"omega/internal/watcher"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

func main() {
	// First, attempt to load the config
	config, err := structs.ConfigFromFile(env.ConfigPath)
	if err != nil {

		// File exists, but cannot seem to load for some reason
		if !errors.Is(err, fs.ErrNotExist) {
			log.Fatal("Error occured loading "+env.ConfigPath, "error", err)
		}

		if !forms.FormGenerateConfig() {
			// Does not want to create config
			// NOTE:: omega does not run without the config file
			log.Info("Alright see ya later 🐊!")
			return
		}

		l := forms.FormSelectLanguage()

		program.SetupLanguage(l)

		config, err = structs.ConfigFromFile(env.ConfigPath)
		if err != nil {
			log.Fatal("Something went wrong reading the config file", "error", err)
		}
		return
	}

	// Config parsed, check validity
	if len(config.Commands) == 0 {
		log.Fatal("No commands specified! Please specify them under the \"commands\" field in omega.json")
	}

	log.SetLogLevel("error")
	for _, logFile := range config.LogFiles {
		dir := filepath.Dir(logFile)
		if dir != "." {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				log.Warn("Unable to create directories for log file skipping", "file", logFile, "error", err)
				continue
			}
		}
		config.Ignore = append(config.Ignore, logFile)
		if filepath.Base(logFile) == logFile {
			config.Ignore = append(config.Ignore, "./"+logFile)
		}
		err = log.InitFileLogger(logFile)
		if err != nil {
			log.Warn("Unable to create specified file logger, skipping", "file", logFile, "error", err)
		}
	}
	log.SetLogLevel(config.LogLevel)
	defer log.CloseAll()

	runnerChannel := make(chan struct{}, 1)
	watcherWorker, err := watcher.New(*config, runnerChannel)
	if err != nil {
		log.Fatal("Error occured creating watcher worker", "error", err)
	}

	runnerWorker := runner.NewRunner(*config, runnerChannel)

	osChannel := make(chan os.Signal, 1)
	signal.Notify(osChannel, syscall.SIGTERM, syscall.SIGINT)

	doneChannel := make(chan error, 2)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		doneChannel <- watcherWorker.Run(ctx)
		waitGroup.Done()
	}()

	go func() {
		doneChannel <- runnerWorker.Run(ctx)
		waitGroup.Done()
	}()

	log.Info("Received OS signal to shutdown", "signal", <-osChannel)
	cancel()
	waitGroup.Wait()
}
