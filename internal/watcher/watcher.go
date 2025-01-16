package watcher

import (
	"context"
	"fmt"
	"io/fs"
	"omega/internal/env"
	"omega/internal/log"
	"omega/internal/structs"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	runnerChannel chan<- struct{}
	watcher       *fsnotify.Watcher

	timeout time.Duration
	ignore  []string
}

func New(config structs.Config, runnerChannel chan<- struct{}) (*Watcher, error) {
	watcher := &Watcher{
		runnerChannel: runnerChannel,
		ignore:        config.Ignore,
		timeout:       time.Duration(config.Timeout) * time.Millisecond,
	}

	fsnWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watcher.watcher = fsnWatcher

	err = watcher.crawl(env.BaseDirectory)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func (watcher *Watcher) walkDirFunc(filePath string, dirEntry fs.DirEntry, _ error) error {
	log.Debug("Found", "file", filePath)

	for i, ignore := range watcher.ignore {
		if ignore == "" {
			continue
		}
		match, err := filepath.Match(ignore, filePath)
		if err != nil {
			log.Warn("Error occured parsing ignore list, skipping", "value", ignore)
			watcher.ignore[i] = ""
			continue
		} else if match {
			log.Debug("Skipping", "file", filePath)
			return fs.SkipDir
		}
	}

	if dirEntry.IsDir() {
		err := watcher.watcher.Add(filePath)
		if err != nil {
			log.Warn("Error occured watching a certain directory", "directory", filePath, "error", err)
		}
	}
	return nil
}

func (watcher *Watcher) crawl(rootDir string) error {
	// fs.WalkDirFunc
	return filepath.WalkDir(rootDir, watcher.walkDirFunc)
}

func (watcher *Watcher) Run(ctx context.Context) error {
	log.Debug("Watcher started")
	defer log.Debug("Watcher shutting down")

	timer := time.NewTimer(0)

mainLoop:
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-timer.C:
			fmt.Print("\033[H\033[2J") // Clear the terminal
			log.Info(env.ProgramName + " reloading...\n")

			// Send signal to reload
			watcher.runnerChannel <- struct{}{}

		case event, ok := <-watcher.watcher.Events:
			if !ok {
				return nil
			}
			extension := filepath.Ext(event.Name)
			log.Debug("event", "name", event.Name,
				"extension", extension,
				"operation", event.Op.String(),
			)
			for i, p := range watcher.ignore {
				if p == "" {
					continue
				}
				match, err := filepath.Match(p, event.Name)
				if err != nil {
					watcher.ignore[i] = ""
				} else if match {
					log.Debug("Skipping event", "name", event.Name)
					continue mainLoop
				}
			}

			timer.Reset(watcher.timeout)
			log.Debug("Watcher timer started", "timeout", watcher.timeout)

		case err, ok := <-watcher.watcher.Errors:
			if !ok {
				return nil
			}
			log.Error("Error occured while watching", "error", err)
		}
	}
}
