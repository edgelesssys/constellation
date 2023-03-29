/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package journald

import (
	"errors"
	"io"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubJournaldCommand struct {
	startError error
}

func (j *stubJournaldCommand) Start() error {
	return j.startError
}

type stubStderrPipe struct {
	buffer   []byte
	readErr  error
	closeErr error
}

func (s stubStderrPipe) Read(p []byte) (n int, err error) {
	size := len(s.buffer)
	if s.readErr != nil {
		size = 0
	}
	return size, s.readErr
}

func (s stubStderrPipe) Close() error {
	return s.closeErr
}

func TestCollect(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		command      *stubJournaldCommand
		wantedOutput []byte
		wantErr      bool
	}{
		"success": {
			command: &stubJournaldCommand{},
		},
		"execution failed": {
			command: &stubJournaldCommand{startError: someError},
			wantErr: true,
		},
		"exit error": {
			command: &stubJournaldCommand{startError: &exec.ExitError{}},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{cmd: tc.command}

			_, err := collector.Pipe()
			if tc.wantErr {
				assert.Error(err)
			}
		})
	}
}

func TestStderr(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		stderrPipe io.ReadCloser
		wantErr    bool
	}{
		"success": {
			stderrPipe: stubStderrPipe{readErr: io.EOF},
		},
		"reading error": {
			stderrPipe: stubStderrPipe{readErr: someError},
			wantErr:    true,
		},
		"close error": {
			stderrPipe: stubStderrPipe{closeErr: someError, readErr: io.EOF},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{
				stderrPipe: tc.stderrPipe,
			}

			stderrOut, err := collector.Stderr()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Equal(stderrOut, []byte{})
			}
		})
	}
}
