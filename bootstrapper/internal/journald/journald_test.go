/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package journald

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipe(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		command      *stubCommand
		stdoutPipe   io.ReadCloser
		wantedOutput []byte
		wantErr      bool
	}{
		"success": {
			command:      &stubCommand{},
			wantedOutput: []byte("asdf"),
			stdoutPipe:   &stubReadCloser{reader: bytes.NewReader([]byte("asdf"))},
		},
		"execution failed": {
			command: &stubCommand{startErr: someError},
			wantErr: true,
		},
		"exit error": {
			command: &stubCommand{startErr: &exec.ExitError{}},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{cmd: tc.command, stdoutPipe: tc.stdoutPipe}

			pipe, err := collector.Start()
			if tc.wantErr {
				assert.Error(err)
			} else {
				stdout := make([]byte, 4)
				_, err = io.ReadFull(pipe, stdout)
				require.NoError(t, err)
				assert.Equal(tc.wantedOutput, stdout)
			}
		})
	}
}

func TestError(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		stderrPipe io.ReadCloser
		exitCode   error
		wantErr    bool
	}{
		"success": {
			stderrPipe: &stubReadCloser{readErr: io.EOF},
		},
		"reading error": {
			stderrPipe: &stubReadCloser{readErr: someError},
			wantErr:    true,
		},
		"close error": {
			stderrPipe: &stubReadCloser{closeErr: someError, readErr: io.EOF},
		},
		"command exit": {
			stderrPipe: &stubReadCloser{readErr: io.EOF},
			exitCode:   someError,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{
				stderrPipe: tc.stderrPipe,
				cmd:        &stubCommand{waitErr: tc.exitCode},
			}

			stderrOut, err := collector.Error()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Equal(stderrOut, []byte{})
			}
		})
	}
}

type stubCommand struct {
	startCalled bool
	startErr    error
	waitErr     error
}

func (j *stubCommand) Start() error {
	j.startCalled = true
	return j.startErr
}

func (j *stubCommand) Wait() error {
	return j.waitErr
}

type stubReadCloser struct {
	reader   io.Reader
	readErr  error
	closeErr error
}

func (s *stubReadCloser) Read(p []byte) (n int, err error) {
	if s.readErr != nil {
		return 0, s.readErr
	}
	return s.reader.Read(p)
}

func (s *stubReadCloser) Close() error {
	return s.closeErr
}
