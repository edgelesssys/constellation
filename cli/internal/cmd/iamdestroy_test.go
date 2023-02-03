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
	"github.com/edgelesssys/constellation/v2/internal/logger"
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
		iamDestroyer      *stubIAMDestroyer
		fh                file.Handler
		stdin             string
		yesFlag           string
		wantErr           bool
		wantDestroyCalled bool
	}{
		"file missing abort": {
			fh:           newFsMissing(),
			stdin:        "n\n",
			yesFlag:      "false",
			iamDestroyer: &stubIAMDestroyer{},
		},
		"file missing": {
			fh:                newFsMissing(),
			stdin:             "y\n",
			yesFlag:           "false",
			iamDestroyer:      &stubIAMDestroyer{},
			wantDestroyCalled: true,
		},
		"file exists abort": {
			fh:           newFsExists(),
			stdin:        "n\n",
			yesFlag:      "false",
			iamDestroyer: &stubIAMDestroyer{},
		},
		"error destroying user": {
			fh:                newFsMissing(),
			stdin:             "y\n",
			yesFlag:           "false",
			iamDestroyer:      &stubIAMDestroyer{destroyErr: someError},
			wantErr:           true,
			wantDestroyCalled: true,
		},
		"gcp delete error": {
			fh:           newFsExists(),
			yesFlag:      "true",
			iamDestroyer: &stubIAMDestroyer{deleteGCPFileErr: someError},
			wantErr:      true,
		},
		"gcp no proceed": {
			fh:           newFsExists(),
			yesFlag:      "true",
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
			assert.NoError(cmd.Flags().Set("yes", tc.yesFlag))

			c := &destroyCmd{log: logger.NewTest(t)}

			err := c.iamDestroy(cmd, &nopSpinner{}, tc.iamDestroyer, tc.fh)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			if tc.wantDestroyCalled {
				assert.True(tc.iamDestroyer.destroyCalled)
			} else {
				assert.False(tc.iamDestroyer.destroyCalled)
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
		destroyer        *stubIAMDestroyer
		fsHandler        file.Handler
		stdin            string
		wantErr          bool
		wantProceed      bool
		wantDeleteCalled bool
	}{
		"file doesn't exist": {
			destroyer:   &stubIAMDestroyer{},
			fsHandler:   newFsNoExist(),
			wantProceed: true,
		},
		"unsuccessful destroy confirm": {
			destroyer:        &stubIAMDestroyer{},
			fsHandler:        newFsExist(),
			stdin:            "y\n",
			wantProceed:      true,
			wantDeleteCalled: true,
		},
		"unsuccessful destroy deny": {
			destroyer:        &stubIAMDestroyer{},
			fsHandler:        newFsExist(),
			stdin:            "n\n",
			wantProceed:      false,
			wantDeleteCalled: true,
		},
		"error deleting file": {
			destroyer:        &stubIAMDestroyer{deleteGCPFileErr: someError},
			fsHandler:        newFsExist(),
			wantErr:          true,
			wantDeleteCalled: true,
		},
		"successful delete": {
			destroyer:        &stubIAMDestroyer{deletedGCPFile: true},
			fsHandler:        newFsExist(),
			wantProceed:      true,
			wantDeleteCalled: true,
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

			if tc.wantDeleteCalled {
				assert.True(tc.destroyer.deleteGCPFileCalled)
			} else {
				assert.False(tc.destroyer.deleteGCPFileCalled)
			}
		})
	}
}
