/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMCreateGCP(t *testing.T) {
	fsWithDefaultConfig := func(require *require.Assertions, provider cloudprovider.Provider) afero.Fs {
		fs := afero.NewMemMapFs()
		file := file.NewHandler(fs)
		require.NoError(file.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), provider)))
		return fs
	}
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: "eyJwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // {"private_key_id":"not_a_secret"}
		},
	}
	invalidIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: "ey_Jwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // invalid b64
		},
	}

	testCases := map[string]struct {
		setupFs              func(*require.Assertions, cloudprovider.Provider) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		zoneFlag             string
		serviceAccountIDFlag string
		projectIDFlag        string
		yesFlag              bool
		stdin                string
		wantAbort            bool
		wantErr              bool
	}{
		"iam create gcp": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
		},
		"iam create gcp invalid flags": {
			setupFs:  fsWithDefaultConfig,
			creator:  &stubIAMCreator{id: validIAMIDFile},
			provider: cloudprovider.GCP,
			zoneFlag: "-a",
			yesFlag:  true,
			wantErr:  true,
		},
		"iam create gcp invalid b64": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: invalidIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			wantErr:              true,
		},
		"interactive": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "yes\n",
		},
		"interactive abort": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "no\n",
			wantAbort:            true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newIAMCreateGCPCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			if tc.zoneFlag != "" {
				require.NoError(cmd.Flags().Set("zone", tc.zoneFlag))
			}
			if tc.serviceAccountIDFlag != "" {
				require.NoError(cmd.Flags().Set("serviceAccountID", tc.serviceAccountIDFlag))
			}
			if tc.projectIDFlag != "" {
				require.NoError(cmd.Flags().Set("projectID", tc.projectIDFlag))
			}
			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider))

			err := iamCreateGCP(cmd, nopSpinner{}, fileHandler, tc.creator)

			if tc.wantErr {
				assert.Error(err)
			} else {
				if tc.wantAbort {
					assert.False(tc.creator.createCalled)
				} else {
					assert.NoError(err)
					assert.True(tc.creator.createCalled)
					assert.Equal(tc.creator.id.GCPOutput, validIAMIDFile.GCPOutput)
				}
			}
		})
	}
}

func TestParseIDFile(t *testing.T) {
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: "eyJwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // {"private_key_id":"not_a_secret"}
		},
	}
	invalidIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: "ey_Jwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // invalid b64
		},
	}
	testCases := map[string]struct {
		idFile           iamid.File
		wantPrivateKeyID string
		wantErr          bool
	}{
		"valid base64": {
			idFile:           validIAMIDFile,
			wantPrivateKeyID: "not_a_secret",
		},
		"invalid base64": {
			idFile:  invalidIAMIDFile,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			outMap, err := parseIDFile(tc.idFile.GCPOutput.ServiceAccountKey)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantPrivateKeyID, outMap["private_key_id"])
			}
		})
	}
}
