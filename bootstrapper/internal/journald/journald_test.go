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
	startError error
}

func (j *stubJournaldCommand) Start() error {
	return j.startError
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
