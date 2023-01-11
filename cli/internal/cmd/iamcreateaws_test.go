/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"strings"
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

func TestIAMCreateAWS(t *testing.T) {
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.AWS,
		AWSOutput: iamid.AWSFile{
			ControlPlaneInstanceProfile: "test_control_plane_instance_profile",
			WorkerNodeInstanceProfile:   "test_worker_nodes_instance_profile",
		},
	}

	testCases := map[string]struct {
		creator    *stubIAMCreator
		provider   cloudprovider.Provider
		zoneFlag   string
		prefixFlag string
		yesFlag    bool
		generateConfigFlag bool
		configFlag string
		existingFiles []string
		stdin      string
		wantAbort  bool
		wantErr    bool
	}{
		"iam create aws": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
		},
		"iam create aws generate config": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
			configFlag: constants.ConfigFilename,
			generateConfigFlag: true,
		},
		"iam create aws generate config custom path": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
			generateConfigFlag: true,
			configFlag: "custom-config.yaml",
		},
		"iam create aws generate config path already exists": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
			generateConfigFlag: true,
			wantErr:   true,
			configFlag: constants.ConfigFilename,
			existingFiles: []string{constants.ConfigFilename},
		},
		"iam create aws generate config custom path already exists": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
			generateConfigFlag: true,
			wantErr:   true,
			configFlag: "custom-config.yaml",
			existingFiles: []string{"custom-config.yaml"},
		},
		"interactive": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "yes\n",
		},
		"interactive generate config": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "yes\n",
			configFlag: constants.ConfigFilename,
			generateConfigFlag: true,
		},
		"interactive abort": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "no\n",
			wantAbort:  true,
		},
		"interactive generate config abort": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "no\n",
			generateConfigFlag: true,
			configFlag: constants.ConfigFilename,
			wantAbort:  true,
		},
		"invalid zone": {
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-west-5b",
			prefixFlag: "test",
			yesFlag:    true,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newIAMCreateAWSCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("generate-config", false, "")             // register persistent flag manually

			if tc.zoneFlag != "" {
				require.NoError(cmd.Flags().Set("zone", tc.zoneFlag))
			}
			if tc.prefixFlag != "" {
				require.NoError(cmd.Flags().Set("prefix", tc.prefixFlag))
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

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			for _, f := range tc.existingFiles {
				require.NoError(fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
			}

			err := iamCreateAWS(cmd, nopSpinner{}, tc.creator, fileHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				if tc.wantAbort {
					assert.False(tc.creator.createCalled)
				} else {
					if tc.generateConfigFlag {
						readConfig := &config.Config{}
						readErr := fileHandler.ReadYAML(tc.configFlag, readConfig)
						require.NoError(readErr)
						assert.Equal(tc.creator.id.AWSOutput.ControlPlaneInstanceProfile, readConfig.Provider.AWS.IAMProfileControlPlane)
						assert.Equal(tc.creator.id.AWSOutput.WorkerNodeInstanceProfile, readConfig.Provider.AWS.IAMProfileWorkerNodes)
						assert.Equal(tc.zoneFlag, readConfig.Provider.AWS.Zone)
						assert.True(strings.HasPrefix(readConfig.Provider.AWS.Zone, readConfig.Provider.AWS.Region))
					}
					require.NoError(err)
					assert.True(tc.creator.createCalled)
					assert.Equal(tc.creator.id.AWSOutput, validIAMIDFile.AWSOutput)
				}
			}
		})
	}
}
