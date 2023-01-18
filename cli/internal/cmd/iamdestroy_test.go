/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestroyIAMUser(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		iamDestroyer iamDestroyer
		stdin        string
		yes          string
		wantErr      bool
	}{
		"confirm okay": {
			iamDestroyer: &stubIAMDestroyer{},
			stdin:        "y\n",
			yes:          "false",
		},
		"confirm abort": {
			iamDestroyer: &stubIAMDestroyer{},
			stdin:        "n\n",
			yes:          "false",
		},
		"destroy fail": {
			iamDestroyer: &stubIAMDestroyer{destroyErr: someError},
			yes:          "true",
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newIAMDestroyCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			cmd.Flags().Set("yes", tc.yes)

			err := destroyIAMUser(cmd, &nopSpinner{}, tc.iamDestroyer)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
