/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package journald

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubJournaldCommand struct {
	executeCommandOutput []byte
	executeCommandError  error
}

func (j *stubJournaldCommand) executeCommand() ([]byte, error) {
	return j.executeCommandOutput, j.executeCommandError
}

func TestGetServiceLog(t *testing.T) {
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
			command: &stubJournaldCommand{executeCommandError: someError},
			wantErr: true,
		},
		"output check": {
			command:      &stubJournaldCommand{executeCommandOutput: []byte("asdf")},
			wantedOutput: []byte("asdf"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			out, err := GetServiceLog(tc.command)
			if tc.wantErr {
				assert.Error(err)
			}
			assert.Equal(out, tc.wantedOutput)
		})
	}
}
