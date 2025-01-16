package main

import (
	"context"
	"omega/internal/log"
	"omega/internal/program"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

func main() {
	c, err := program.Setup()
	if err != nil {
		log.Fatal(err)
	}

	// Config parsed, check validity
	if len(c.Commands) == 0 {
		log.Warn("No commands specified! Please specify them under the \"commands\" field in omega.json")
	}

	log.SetLogLevel("error")
	for _, logFile := range c.LogFiles {
		dir := filepath.Dir(logFile)
		if dir != "." {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				log.Warn("Unable to create directories for log file skipping", "file", logFile, "error", err)
				continue
			}
		}
		c.Ignore = append(c.Ignore, logFile)
		if filepath.Base(logFile) == logFile {
			c.Ignore = append(c.Ignore, "./"+logFile)
		}
		err = log.NewFromFile(logFile)
		if err != nil {
			log.Warn("Unable to create specified file logger, skipping", "file", logFile, "error", err)
		}
	}
	log.SetLogLevel(c.LogLevel)
	defer log.CloseAll()

	workerCh := make(chan struct{}, 1)
	watcher, err := program.NewWatcher(*c, workerCh)
	if err != nil {
		log.Fatal("Error occured creating watcher worker", "error", err)
	}

	runner := program.NewRunner(*c, workerCh)

	osCh := make(chan os.Signal, 1)
	signal.Notify(osCh, syscall.SIGTERM, syscall.SIGINT)

	doneCh := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		doneCh <- watcher.Start(ctx)
		wg.Done()
	}()

	go func() {
		doneCh <- runner.Start(ctx)
		wg.Done()
	}()
	select {
	case err = <-doneCh:
		log.Error("Error occurred by one of the workers", "error", err)
	case sig := <-osCh:
		log.Info("Received OS signal to shutdown", "signal", sig)
	}
	cancel()
	wg.Wait()
}
