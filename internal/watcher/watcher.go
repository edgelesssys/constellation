/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

import (
	"fmt"

	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// FileWatcher watches for changes to the file and calls the waiter's Update method.
type FileWatcher struct {
	log     *logger.Logger
	updater updater
	watcher eventWatcher
	done    chan struct{}
}

// New creates a new FileWatcher for the given validator.
func New(log *logger.Logger, updater updater) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		log:     log,
		watcher: &fsnotifyWatcher{watcher},
		updater: updater,
		done:    make(chan struct{}, 1),
	}, nil
}

// Close closes the watcher.
// It should only be called once.
func (f *FileWatcher) Close() error {
	err := f.watcher.Close()
	<-f.done
	return err
}

// Watch starts watching the file at the given path.
// It will call the watcher's Update method when the file is modified.
func (f *FileWatcher) Watch(file string) error {
	log := f.log.With("file", file)
	defer func() { f.done <- struct{}{} }()
	if err := f.watcher.Add(file); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-f.watcher.Events():
			if !ok {
				log.Infof("Watcher closed")
				return nil
			}

			// file changes may be indicated by either a WRITE, CHMOD, CREATE or RENAME event
			if event.Op&(fsnotify.Write|fsnotify.Chmod|fsnotify.Create|fsnotify.Rename) != 0 {
				if err := f.updater.Update(); err != nil {
					log.With(zap.Error(err)).Errorf("Update failed")
				}
			}

			// if a file gets removed, e.g. by a rename event, we need to re-add the file to the watcher
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				if err := f.watcher.Add(event.Name); err != nil {
					log.With(zap.Error(err)).Errorf("Failed to re-add file to watcher")
					return fmt.Errorf("failed to re-add file %q to watcher: %w", event.Name, err)
				}
			}

		case err := <-f.watcher.Errors():
			if err != nil {
				log.With(zap.Error(err)).Errorf("Watching for measurements updates")
				return fmt.Errorf("watching for measurements updates: %w", err)
			}
		}
	}
}

type updater interface {
	Update() error
}

type eventWatcher interface {
	Add(string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

type fsnotifyWatcher struct {
	watcher *fsnotify.Watcher
}

func (w *fsnotifyWatcher) Add(file string) error {
	return w.watcher.Add(file)
}

func (w *fsnotifyWatcher) Close() error {
	return w.watcher.Close()
}

func (w *fsnotifyWatcher) Events() <-chan fsnotify.Event {
	return w.watcher.Events
}

func (w *fsnotifyWatcher) Errors() <-chan error {
	return w.watcher.Errors
}
