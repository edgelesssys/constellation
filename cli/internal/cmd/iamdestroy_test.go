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

func TestIAMDestroy(t *testing.T) {
	require := require.New(t)
	someError := errors.New("failed")

	newFsExists := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fh.Write(constants.GCPServiceAccountKeyFile, []byte("{}")))
		return fh
	}
	newFsMissing := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		return fh
	}

	testCases := map[string]struct {
		iamDestroyer iamDestroyer
		fh           file.Handler
		stdin        string
		yes          string
		wantErr      bool
	}{
		"file missing abort": {
			fh:    newFsMissing(),
			stdin: "n\n",
			yes:   "false",
		},
		"file missing": {
			fh:           newFsMissing(),
			stdin:        "y\n",
			yes:          "false",
			iamDestroyer: &stubIAMDestroyer{},
		},
		"file exists abort": {
			fh:    newFsExists(),
			stdin: "n\n",
			yes:   "false",
		},
		"error destroying user": {
			fh:           newFsMissing(),
			stdin:        "y\n",
			yes:          "false",
			iamDestroyer: &stubIAMDestroyer{destroyErr: someError},
			wantErr:      true,
		},
		"gcp delete error": {
			fh:           newFsExists(),
			yes:          "true",
			iamDestroyer: &stubIAMDestroyer{deleteGCPFileErr: someError},
			wantErr:      true,
		},
		"gcp no proceed": {
			fh:           newFsExists(),
			yes:          "true",
			stdin:        "n\n",
			iamDestroyer: &stubIAMDestroyer{deletedGCPFile: false},
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

			err := iamDestroy(cmd, &nopSpinner{}, tc.iamDestroyer, tc.fh)

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

	newFsNoExist := func() file.Handler {
		fs := file.NewHandler(afero.NewMemMapFs())
		return fs
	}
	newFsExist := func() file.Handler {
		fs := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fs.Write(constants.GCPServiceAccountKeyFile, []byte("{}")))
		return fs
	}

	testCases := map[string]struct {
		destroyer   iamDestroyer
		fsHandler   file.Handler
		stdin       string
		wantErr     bool
		wantProceed bool
	}{
		"file doesn't exist": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   newFsNoExist(),
			wantProceed: true,
			wantErr:     true,
		},
		"unsuccessful destroy confirm": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   newFsExist(),
			stdin:       "y\n",
			wantProceed: true,
		},
		"unsuccessful destroy deny": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   newFsExist(),
			stdin:       "n\n",
			wantProceed: false,
		},
		"error deleting file": {
			destroyer: &stubIAMDestroyer{deleteGCPFileErr: someError},
			fsHandler: newFsExist(),
			wantErr:   true,
		},
		"successful delete": {
			destroyer:   &stubIAMDestroyer{deletedGCPFile: true},
			fsHandler:   newFsExist(),
			wantProceed: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newIAMDestroyCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

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
