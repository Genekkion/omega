package program

import (
	"context"
	"io/fs"
	"omega/internal/config"
	"omega/internal/env"
	"omega/internal/log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	ch chan<- struct{}
	w  *fsnotify.Watcher

	timeout time.Duration
	ignore  []string
}

func NewWatcher(config config.Config, ch chan<- struct{}) (*Watcher, error) {
	w := &Watcher{
		ch:      ch,
		ignore:  config.Ignore,
		timeout: time.Duration(config.Timeout) * time.Millisecond,
	}

	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w.w = fsn

	err = w.crawl(env.BaseDirectory)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) walkDirFunc(path string, dirEntry fs.DirEntry, _ error) error {
	log.Debug("Found", "file", path)

	for i, ignore := range w.ignore {
		if ignore == "" {
			continue
		}
		match, err := filepath.Match(ignore, path)
		if err != nil {
			log.Warn("Error occured parsing ignore list, skipping", "value", ignore)
			w.ignore[i] = ""
			continue
		} else if match {
			log.Debug("Skipping", "file", path)
			return fs.SkipDir
		}
	}

	if dirEntry.IsDir() {
		err := w.w.Add(path)
		if err != nil {
			log.Warn("Error occured watching a certain directory", "directory", path, "error", err)
		}
	}
	return nil
}

func (w *Watcher) crawl(rootDir string) error {
	// fs.WalkDirFunc
	return filepath.WalkDir(rootDir, w.walkDirFunc)
}

func (w *Watcher) Start(ctx context.Context) error {
	log.Debug("Watcher started")
	defer log.Debug("Watcher shutting down")

	timer := time.NewTimer(0)

mainLoop:
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-timer.C:
			// fmt.Print("\033[H\033[2J") // Clear the terminal
			log.Info(env.ProgramName + " reloading...\n")

			// Send signal to reload
			w.ch <- struct{}{}

		case event, ok := <-w.w.Events:
			// Pipe closed for some reason
			if !ok {
				return nil
			}

			{
				extension := filepath.Ext(event.Name)
				log.Debug("event", "name", event.Name,
					"extension", extension,
					"operation", event.Op.String(),
				)
			}
			name := strings.TrimPrefix(event.Name, "./")

			for i, p := range w.ignore {
				if p == "" {
					continue
				}
				match, err := filepath.Match(p, name)
				if err != nil {
					w.ignore[i] = ""
				} else if match {
					log.Debug("Skipping event", "name", name)
					continue mainLoop
				}
			}

			timer.Reset(w.timeout)
			log.Debug("Watcher timer started", "timeout", w.timeout)

		case err, ok := <-w.w.Errors:
			if !ok {
				return nil
			}
			log.Error("Error occured while watching", "error", err)
		}
	}
}
