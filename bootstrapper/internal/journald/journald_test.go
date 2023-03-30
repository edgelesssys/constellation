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
)

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
	readErr  error
	reader   io.Reader
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

func TestPipe(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		cmd          *stubCommand
		stdoutPipe   io.ReadCloser
		wantedOutput []byte
		wantErr      bool
	}{
		"success": {
			cmd:          &stubCommand{},
			stdoutPipe:   &stubReadCloser{reader: bytes.NewReader([]byte("asdf"))},
			wantedOutput: []byte("asdf"),
		},
		"execution failed": {
			cmd:     &stubCommand{startErr: someError},
			wantErr: true,
		},
		"exit error": {
			cmd:     &stubCommand{startErr: &exec.ExitError{}},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{
				cmd:    tc.cmd,
				stdout: tc.stdoutPipe,
			}

			pipe, err := collector.Pipe()

			if tc.wantErr {
				assert.Error(err)
			} else {
				out, err := io.ReadAll(pipe)
				assert.NoError(err)
				assert.Equal(tc.wantedOutput, out)
			}
		})
	}
}

func TestError(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		stderrPipe io.ReadCloser
		cmd        *stubCommand
		wantErr    bool
	}{
		"success": {
			stderrPipe: &stubReadCloser{readErr: io.EOF},
			cmd:        &stubCommand{},
		},
		"reading error": {
			stderrPipe: &stubReadCloser{readErr: someError},
			cmd:        &stubCommand{},
			wantErr:    true,
		},
		"close error": {
			stderrPipe: &stubReadCloser{readErr: io.EOF, closeErr: someError},
			cmd:        &stubCommand{},
		},
		"command exit": {
			stderrPipe: &stubReadCloser{readErr: io.EOF},
			cmd:        &stubCommand{waitErr: &exec.ExitError{}},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{
				stderr: tc.stderrPipe,
				cmd:    tc.cmd,
			}

			stderrOut, err := collector.Error()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(stderrOut, []byte{})
			}
		})
	}
}
