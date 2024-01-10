/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"testing"
  "log/slog"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckLicense(t *testing.T) {
	testCases := map[string]struct {
		licenseChecker *stubLicenseChecker
		wantErr        bool
	}{
		"success": {
			licenseChecker: &stubLicenseChecker{},
			wantErr:        false,
		},
		"check license error": {
			licenseChecker: &stubLicenseChecker{checkLicenseErr: assert.AnError},
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			a := &Applier{licenseChecker: tc.licenseChecker, log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil))}
			_, err := a.CheckLicense(context.Background(), cloudprovider.Unknown, true, license.CommunityLicense)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

type stubLicenseChecker struct {
	checkLicenseErr error
}

func (c *stubLicenseChecker) CheckLicense(context.Context, cloudprovider.Provider, license.Action, string) (int, error) {
	return 0, c.checkLicenseErr
}

func TestGenerateMasterSecret(t *testing.T) {
	assert := assert.New(t)
	a := &Applier{log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil))}
	sec, err := a.GenerateMasterSecret()
	assert.NoError(err)
	assert.Len(sec.Key, crypto.MasterSecretLengthDefault)
	assert.Len(sec.Key, crypto.RNGLengthDefault)
}

func TestGenerateMeasurementSalt(t *testing.T) {
	assert := assert.New(t)
	a := &Applier{log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil))}
	salt, err := a.GenerateMeasurementSalt()
	assert.NoError(err)
	assert.Len(salt, crypto.RNGLengthDefault)
}
