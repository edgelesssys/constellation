/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
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
		require.NoError(fh.Write(constants.GCPServiceAccountKeyFilename, []byte("{}")))
		return fh
	}
	newFsMissing := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		return fh
	}
	newFsWithAdminConf := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fh.Write(constants.AdminConfFilename, []byte("")))
		return fh
	}
	newFsWithStateFile := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fh.Write(constants.StateFilename, []byte("")))
		return fh
	}

	testCases := map[string]struct {
		iamDestroyer      *stubIAMDestroyer
		fh                file.Handler
		stdin             string
		yesFlag           bool
		wantErr           bool
		wantDestroyCalled bool
	}{
		"cluster running admin conf": {
			fh:           newFsWithAdminConf(),
			iamDestroyer: &stubIAMDestroyer{},
			yesFlag:      false,
			wantErr:      true,
		},
		"cluster running cluster state": {
			fh:           newFsWithStateFile(),
			iamDestroyer: &stubIAMDestroyer{},
			yesFlag:      false,
			wantErr:      true,
		},
		"file missing abort": {
			fh:           newFsMissing(),
			stdin:        "n\n",
			yesFlag:      false,
			iamDestroyer: &stubIAMDestroyer{},
		},
		"file missing": {
			fh:                newFsMissing(),
			stdin:             "y\n",
			yesFlag:           false,
			iamDestroyer:      &stubIAMDestroyer{},
			wantDestroyCalled: true,
		},
		"file exists abort": {
			fh:           newFsExists(),
			stdin:        "n\n",
			yesFlag:      false,
			iamDestroyer: &stubIAMDestroyer{},
		},
		"error destroying user": {
			fh:                newFsMissing(),
			stdin:             "y\n",
			yesFlag:           false,
			iamDestroyer:      &stubIAMDestroyer{destroyErr: someError},
			wantErr:           true,
			wantDestroyCalled: true,
		},
		"gcp delete error": {
			fh:           newFsExists(),
			yesFlag:      true,
			iamDestroyer: &stubIAMDestroyer{getTfStateKeyErr: someError},
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

      c := &destroyCmd{log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), flags: iamDestroyFlags{
				yes: tc.yesFlag,
			}}

			err := c.iamDestroy(cmd, &nopSpinner{}, tc.iamDestroyer, tc.fh)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.wantDestroyCalled, tc.iamDestroyer.destroyCalled)
		})
	}
}

func TestDeleteGCPServiceAccountKeyFile(t *testing.T) {
	require := require.New(t)
	someError := errors.New("failed")

	gcpFile := `
	{
		"auth_provider_x509_cert_url": "",
		"auth_uri": "",
		"client_email": "",
		"client_id": "",
		"client_x509_cert_url": "",
		"private_key": "",
		"private_key_id": "",
		"project_id": "",
		"token_uri": "",
		"type": ""
	}
	`

	newFs := func() file.Handler {
		fs := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fs.Write(constants.GCPServiceAccountKeyFilename, []byte(gcpFile)))
		return fs
	}
	newFsInvalidJSON := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fh.Write(constants.GCPServiceAccountKeyFilename, []byte("asdf")))
		return fh
	}

	testCases := map[string]struct {
		destroyer          *stubIAMDestroyer
		fsHandler          file.Handler
		stdin              string
		wantErr            bool
		wantProceed        bool
		wantGetSaKeyCalled bool
	}{
		"invalid gcp json": {
			destroyer: &stubIAMDestroyer{},
			fsHandler: newFsInvalidJSON(),
			wantErr:   true,
		},
		"error getting key terraform": {
			destroyer:          &stubIAMDestroyer{getTfStateKeyErr: someError},
			fsHandler:          newFs(),
			wantErr:            true,
			wantGetSaKeyCalled: true,
		},
		"keys not same": {
			destroyer: &stubIAMDestroyer{gcpSaKey: gcpshared.ServiceAccountKey{
				Type: "somethingelse",
			}},
			fsHandler:          newFs(),
			wantGetSaKeyCalled: true,
			wantProceed:        true,
		},
		"valid": {
			destroyer:          &stubIAMDestroyer{},
			fsHandler:          newFs(),
			wantGetSaKeyCalled: true,
			wantProceed:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newIAMDestroyCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

      c := &destroyCmd{log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil))}

			proceed, err := c.deleteGCPServiceAccountKeyFile(cmd, tc.destroyer, tc.fsHandler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			assert.Equal(tc.wantProceed, proceed)
			assert.Equal(tc.wantGetSaKeyCalled, tc.destroyer.getTfStateKeyCalled)
		})
	}
}
