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
	defaultFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingFiles {
			require.NoError(fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
		}
		return fs
	}
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string) afero.Fs {
		fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
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
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		zoneFlag             string
		serviceAccountIDFlag string
		projectIDFlag        string
		yesFlag              bool
		generateConfigFlag   bool
		configFlag           string
		existingFiles        []string
		stdin                string
		wantAbort            bool
		wantErr              bool
	}{
		"iam create gcp": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
		},
		"iam create gcp generate config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			generateConfigFlag:   true,
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
		},
		"iam create gcp generate config custom path": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			generateConfigFlag:   true,
			configFlag:           "custom-config.yaml",
			yesFlag:              true,
		},
		"iam create gcp generate config path already exists": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			generateConfigFlag:   true,
			configFlag:           constants.ConfigFilename,
			existingFiles:        []string{constants.ConfigFilename},
			yesFlag:              true,
			wantErr:              true,
		},
		"iam create gcp generate config custom path already exists": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			generateConfigFlag:   true,
			configFlag:           "custom-config.yaml",
			existingFiles:        []string{"custom-config.yaml"},
			yesFlag:              true,
			wantErr:              true,
		},
		"iam create gcp invalid flags": {
			setupFs:  defaultFs,
			creator:  &stubIAMCreator{id: validIAMIDFile},
			provider: cloudprovider.GCP,
			zoneFlag: "-a",
			yesFlag:  true,
			wantErr:  true,
		},
		"iam create gcp invalid b64": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: invalidIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			wantErr:              true,
		},
		"interactive": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "yes\n",
		},
		"interactive generate config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "yes\n",
			configFlag:           constants.ConfigFilename,
			generateConfigFlag:   true,
		},
		"interactive abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "no\n",
			wantAbort:            true,
		},
		"interactive abort generate config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "no\n",
			wantAbort:            true,
			configFlag:           constants.ConfigFilename,
			generateConfigFlag:   true,
		},
		"unwritable fs": {
			setupFs:              readOnlyFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			generateConfigFlag:   true,
			configFlag:           constants.ConfigFilename,
			wantErr:              true,
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

			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("generate-config", false, "")             // register persistent flag manually

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
			if tc.generateConfigFlag {
				require.NoError(cmd.Flags().Set("generate-config", "true"))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingFiles))

			err := iamCreateGCP(cmd, nopSpinner{}, tc.creator, fileHandler)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			if tc.wantAbort {
				assert.False(tc.creator.createCalled)
				return
			}

			if tc.generateConfigFlag {
				readConfig := &config.Config{}
				readErr := fileHandler.ReadYAML(tc.configFlag, readConfig)
				require.NoError(readErr)
				assert.Equal(constants.GCPServiceAccountKeyFile, readConfig.Provider.GCP.ServiceAccountKeyPath)
			}
			require.NoError(err)
			assert.True(tc.creator.createCalled)
			assert.Equal(tc.creator.id.GCPOutput, validIAMIDFile.GCPOutput)
			readServiceAccountKey := &map[string]string{}
			readErr := fileHandler.ReadJSON(constants.GCPServiceAccountKeyFile, readServiceAccountKey)
			require.NoError(readErr)
			assert.Equal("not_a_secret", (*readServiceAccountKey)["private_key_id"])
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
