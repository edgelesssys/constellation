/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package journald

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubJournaldCommand struct {
	OutputReturn []byte
	OutputError  error
}

func (j *stubJournaldCommand) Output() ([]byte, error) {
	return j.OutputReturn, j.OutputError
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
			command: &stubJournaldCommand{OutputError: someError},
			wantErr: true,
		},
		"exit error": {
			command: &stubJournaldCommand{OutputError: &exec.ExitError{}},
			wantErr: true,
		},
		"output check": {
			command:      &stubJournaldCommand{OutputReturn: []byte("asdf")},
			wantedOutput: []byte("asdf"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			collector := Collector{cmd: tc.command}

			out, err := collector.Collect()
			if tc.wantErr {
				assert.Error(err)
			}
			assert.Equal(out, tc.wantedOutput)
		})
	}
}
