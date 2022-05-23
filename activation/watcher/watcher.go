package watcher

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

// FileWatcher watches for changes to the file and calls the waiter's Update method.
type FileWatcher struct {
	updater updater
	watcher eventWatcher
	done    chan struct{}
}

// New creates a new FileWatcher for the given validator.
func New(updater updater) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
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
	defer func() { f.done <- struct{}{} }()
	if err := f.watcher.Add(file); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-f.watcher.Events():
			if !ok {
				klog.V(4).Infof("watcher closed")
				return nil
			}

			// file changes may be indicated by either a WRITE, CHMOD, CREATE or RENAME event
			if event.Op&(fsnotify.Write|fsnotify.Chmod|fsnotify.Create|fsnotify.Rename) != 0 {
				if err := f.updater.Update(); err != nil {
					klog.Errorf("failed to update activation validator: %s", err)
				}
			}

			// if a file gets removed, e.g. by a rename event, we need to re-add the file to the watcher
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				if err := f.watcher.Add(event.Name); err != nil {
					klog.Errorf("failed to re-add file %q to watcher: %s", event.Name, err)
					return fmt.Errorf("failed to re-add file %q to watcher: %w", event.Name, err)
				}
			}

		case err := <-f.watcher.Errors():
			if err != nil {
				klog.Errorf("watching for measurements updates: %s", err)
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
