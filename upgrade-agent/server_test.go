/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgradeagent

import (
	"context"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/upgrade-agent/upgradeproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionVerifier(t *testing.T) {
	testCases := map[string]struct {
		versionString string
		wantErr       bool
	}{
		"valid version": {
			versionString: "v1.1.1",
		},
		"v prefix missing": {
			versionString: "1.1.1",
			wantErr:       true,
		},
		"invalid space": {
			versionString: "v 1.1.1",
			wantErr:       true,
		},
		"invalid version": {
			versionString: "v1.1.1a",
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			err := verifyVersion(tc.versionString)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
		})
	}
}

func TestPrepareUpdate(t *testing.T) {
	validUpdateRequest := &upgradeproto.ExecuteUpdateRequest{
		WantedKubernetesVersion: "v1.1.1",
	}
	testCases := map[string]struct {
		installer     osInstaller
		updateRequest *upgradeproto.ExecuteUpdateRequest
		wantErr       bool
	}{
		"works": {
			installer:     stubOsInstaller{},
			updateRequest: validUpdateRequest,
		},
		"invalid version string": {
			installer:     stubOsInstaller{},
			updateRequest: &upgradeproto.ExecuteUpdateRequest{WantedKubernetesVersion: "1337"},
			wantErr:       true,
		},
		"install error": {
			installer:     stubOsInstaller{InstallErr: fmt.Errorf("install error")},
			updateRequest: validUpdateRequest,
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			err := prepareUpdate(context.Background(), tc.installer, tc.updateRequest)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
		})
	}
}

type stubOsInstaller struct {
	InstallErr error
}

func (s stubOsInstaller) Install(ctx context.Context, kubernetesComponent versions.ComponentVersion) error {
	return s.InstallErr
}
