/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

import (
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestWatcher(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		updater      *testUpdater
		watcher      *testWatcher
		events       []fsnotify.Event
		watchErr     error
		wantAddCalls int
		wantErr      bool
	}{
		"success": {
			updater: &testUpdater{},
			watcher: &testWatcher{
				events: make(chan fsnotify.Event, 1),
				errors: make(chan error, 1),
			},
			events: []fsnotify.Event{
				{Op: fsnotify.Write, Name: "test"},
				{Op: fsnotify.Chmod, Name: "test"},
				{Op: fsnotify.Create, Name: "test"},
				{Op: fsnotify.Rename, Name: "test"},
			},
			wantAddCalls: 1,
		},
		"failing update does not interrupt execution": {
			updater: &testUpdater{
				err: someErr,
			},
			watcher: &testWatcher{
				events: make(chan fsnotify.Event, 1),
				errors: make(chan error, 1),
			},
			events: []fsnotify.Event{
				{Op: fsnotify.Write, Name: "test"},
				{Op: fsnotify.Write, Name: "test"},
			},
			wantAddCalls: 1,
		},
		"removed file gets re-added": {
			updater: &testUpdater{},
			watcher: &testWatcher{
				events: make(chan fsnotify.Event, 1),
				errors: make(chan error, 1),
			},
			events: []fsnotify.Event{
				{Op: fsnotify.Write, Name: "test"},
				{Op: fsnotify.Remove, Name: "test"},
			},
			wantAddCalls: 2,
		},
		"re-adding file fails": {
			updater: &testUpdater{},
			watcher: &testWatcher{
				addErr: someErr,
				events: make(chan fsnotify.Event, 1),
				errors: make(chan error, 1),
			},
			events:       []fsnotify.Event{{Op: fsnotify.Remove, Name: "test"}},
			wantAddCalls: 1,
			wantErr:      true,
		},
		"add file fails": {
			updater: &testUpdater{},
			watcher: &testWatcher{
				addErr: someErr,
				events: make(chan fsnotify.Event, 1),
				errors: make(chan error, 1),
			},
			wantAddCalls: 1,
			wantErr:      true,
		},
		"error during watch": {
			updater: &testUpdater{},
			watcher: &testWatcher{
				events: make(chan fsnotify.Event, 1),
				errors: make(chan error, 1),
			},
			wantAddCalls: 1,
			watchErr:     someErr,
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			watcher := &FileWatcher{
        log:     slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				updater: tc.updater,
				watcher: tc.watcher,
				done:    make(chan struct{}, 1),
			}

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := watcher.Watch("test")
				if tc.wantErr {
					assert.Error(err)
					return
				}
				assert.NoError(err)
			}()

			time.Sleep(15 * time.Millisecond)

			for _, event := range tc.events {
				tc.watcher.events <- event
			}

			if tc.watchErr != nil {
				tc.watcher.errors <- tc.watchErr
			}

			close(tc.watcher.events)
			assert.NoError(watcher.Close())
			wg.Wait()

			// check that the watchers Add method was called the expected number of times
			assert.Equal(tc.wantAddCalls, tc.watcher.addCalled)
		})
	}
}

type testUpdater struct {
	err error
}

func (u *testUpdater) Update() error {
	return u.err
}

type testWatcher struct {
	addCalled int
	addErr    error
	closeErr  error
	events    chan fsnotify.Event
	errors    chan error
}

func (w *testWatcher) Add(_ string) error {
	w.addCalled++
	return w.addErr
}

func (w *testWatcher) Close() error {
	return w.closeErr
}

func (w *testWatcher) Events() <-chan fsnotify.Event {
	return w.events
}

func (w *testWatcher) Errors() <-chan error {
	return w.errors
}
