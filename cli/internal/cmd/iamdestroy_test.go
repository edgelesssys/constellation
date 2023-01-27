/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.NoError(cmd.Flags().Set("yes", tc.yes))

			fsh := file.NewHandler(afero.NewMemMapFs())

			err := destroyIAMUser(cmd, &nopSpinner{}, tc.iamDestroyer, fsh)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestDeleteGCPServiceAccountKeyFile(t *testing.T) {
	require := require.New(t)
	someError := errors.New("failed")

	fsExist := file.NewHandler(afero.NewMemMapFs())
	fsNoExist := file.NewHandler(afero.NewMemMapFs())
	require.NoError(fsExist.Write(constants.GCPServiceAccountKeyFile, []byte("{}")))

	testCases := map[string]struct {
		destroyer   iamDestroyer
		fsHandler   file.Handler
		yes         string
		stdin       string
		wantErr     bool
		wantProceed bool
	}{
		"file doesn't exist": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   fsNoExist,
			wantProceed: true,
			wantErr:     true,
			yes:         "false",
		},
		"confirm delete flag": {
			destroyer:   &stubIAMDestroyer{deletedGCPFile: true},
			fsHandler:   fsExist,
			wantProceed: true,
			yes:         "true",
		},
		"confirm delete stdin": {
			destroyer:   &stubIAMDestroyer{deletedGCPFile: true},
			fsHandler:   fsExist,
			wantProceed: true,
			yes:         "false",
			stdin:       "y\n",
		},
		"deny delete stdin": {
			destroyer:   &stubIAMDestroyer{deletedGCPFile: true},
			fsHandler:   fsExist,
			wantProceed: true,
			yes:         "false",
			stdin:       "n\n",
		},
		"unsuccessful destroy confirm": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   fsExist,
			yes:         "true",
			stdin:       "y\n",
			wantProceed: true,
		},
		"unsuccessful destroy deny": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   fsExist,
			yes:         "true",
			stdin:       "n\n",
			wantProceed: false,
		},
		"error deleting file": {
			destroyer: &stubIAMDestroyer{deleteGCPFileErr: someError},
			fsHandler: fsExist,
			yes:       "true",
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newIAMDestroyCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			assert.NoError(cmd.Flags().Set("yes", tc.yes))

			proceed, err := deleteGCPServiceAccountKeyFile(cmd, tc.destroyer, tc.fsHandler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.wantProceed {
				assert.True(proceed)
			} else {
				assert.False(proceed)
			}
		})
	}
}
