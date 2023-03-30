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
	"github.com/stretchr/testify/require"
)

type stubCommand struct {
	stdout     *stubStdoutPipe
	text       []byte
	startError error
	exitCode   error
}

func (j *stubCommand) Start() error {
	j.stdout.buffer = j.text
	return j.startError
}

func (j *stubCommand) Wait() error {
	return j.exitCode
}

type stubStdoutPipe struct {
	buffer []byte
	read   bool
}

func (s *stubStdoutPipe) Read(p []byte) (int, error) {
	if !s.read {
		s.read = true
		for i := range p {
			p[i] = s.buffer[i]
		}
		return len(p), nil
	}
	return 0, io.EOF
}

func (s stubStdoutPipe) Close() error {
	return nil
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

func TestPipe(t *testing.T) {
	someError := errors.New("failed")
	stdoutPipe := stubStdoutPipe{}

	testCases := map[string]struct {
		command      *stubCommand
		stdoutPipe   io.ReadCloser
		wantedOutput []byte
		wantErr      bool
	}{
		"success": {
			command:      &stubCommand{stdout: &stdoutPipe, text: []byte("asdf")},
			wantedOutput: []byte("asdf"),
			stdoutPipe:   &stdoutPipe,
		},
		"execution failed": {
			command: &stubCommand{startError: someError, stdout: &stdoutPipe},
			wantErr: true,
		},
		"exit error": {
			command: &stubCommand{startError: &exec.ExitError{}, stdout: &stdoutPipe},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{cmd: tc.command, stdoutPipe: &tc.stdoutPipe}

			pipe, err := collector.Pipe()
			if tc.wantErr {
				assert.Error(err)
			} else {
				stdout := make([]byte, 4)
				_, err = io.ReadFull(*pipe, stdout)
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
			stderrPipe: stubStderrPipe{readErr: io.EOF},
		},
		"reading error": {
			stderrPipe: stubStderrPipe{readErr: someError},
			wantErr:    true,
		},
		"close error": {
			stderrPipe: stubStderrPipe{closeErr: someError, readErr: io.EOF},
		},
		"command exit": {
			stderrPipe: stubStderrPipe{readErr: io.EOF},
			exitCode:   someError,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{
				stderrPipe: &tc.stderrPipe,
				cmd:        &stubCommand{exitCode: tc.exitCode},
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
