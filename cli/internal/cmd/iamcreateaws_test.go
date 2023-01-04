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

func TestIAMCreateAWS(t *testing.T) {
	fsWithDefaultConfig := func(require *require.Assertions, provider cloudprovider.Provider) afero.Fs {
		fs := afero.NewMemMapFs()
		file := file.NewHandler(fs)
		require.NoError(file.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), provider)))
		return fs
	}
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.AWS,
		AWSOutput: iamid.AWSFile{
			ControlPlaneInstanceProfile: "test_control_plane_instance_profile",
			WorkerNodeInstanceProfile:   "test_worker_nodes_instance_profile",
		},
	}

	testCases := map[string]struct {
		setupFs    func(*require.Assertions, cloudprovider.Provider) afero.Fs
		creator    *stubIAMCreator
		provider   cloudprovider.Provider
		zoneFlag   string
		prefixFlag string
		yesFlag    bool
		stdin      string
		wantAbort  bool
		wantErr    bool
	}{
		"iam create aws": {
			setupFs:    fsWithDefaultConfig,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
		},
		"interactive": {
			setupFs:    fsWithDefaultConfig,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "yes\n",
		},
		"interactive abort": {
			setupFs:    fsWithDefaultConfig,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "no\n",
			wantAbort:  true,
		},
		"invalid zone": {
			setupFs:    fsWithDefaultConfig,
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
			if tc.zoneFlag != "" {
				require.NoError(cmd.Flags().Set("zone", tc.zoneFlag))
			}
			if tc.prefixFlag != "" {
				require.NoError(cmd.Flags().Set("prefix", tc.prefixFlag))
			}
			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}

			err := iamCreateAWS(cmd, &nopSpinner{}, tc.creator)

			if tc.wantErr {
				assert.Error(err)
			} else {
				if tc.wantAbort {
					assert.False(tc.creator.createCalled)
				} else {
					assert.NoError(err)
					assert.True(tc.creator.createCalled)
					assert.Equal(tc.creator.id.AWSOutput, validIAMIDFile.AWSOutput)
				}
			}
		})
	}
}
