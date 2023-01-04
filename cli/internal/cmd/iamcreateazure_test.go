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

func TestIAMCreateAzure(t *testing.T) {
	fsWithDefaultConfig := func(require *require.Assertions, provider cloudprovider.Provider) afero.Fs {
		fs := afero.NewMemMapFs()
		file := file.NewHandler(fs)
		require.NoError(file.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), provider)))
		return fs
	}
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: iamid.AzureFile{
			SubscriptionID:               "test_subscription_id",
			TenantID:                     "test_tenant_id",
			ApplicationID:                "test_application_id",
			ApplicationClientSecretValue: "test_application_client_secret_value",
			UAMIID:                       "test_uami_id",
		},
	}

	testCases := map[string]struct {
		setupFs              func(*require.Assertions, cloudprovider.Provider) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		regionFlag           string
		servicePrincipalFlag string
		resourceGroupFlag    string
		yesFlag              bool
		stdin                string
		wantAbort            bool
		wantErr              bool
	}{
		"iam create azure": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
		},
		"interactive": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "yes\n",
		},
		"interactive abort": {
			setupFs:              fsWithDefaultConfig,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "no\n",
			wantAbort:            true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newIAMCreateAzureCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			if tc.regionFlag != "" {
				require.NoError(cmd.Flags().Set("region", tc.regionFlag))
			}
			if tc.resourceGroupFlag != "" {
				require.NoError(cmd.Flags().Set("resourceGroup", tc.resourceGroupFlag))
			}
			if tc.servicePrincipalFlag != "" {
				require.NoError(cmd.Flags().Set("servicePrincipal", tc.servicePrincipalFlag))
			}
			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}

			err := iamCreateAzure(cmd, &nopSpinner{}, tc.creator)

			if tc.wantErr {
				assert.Error(err)
			} else {
				if tc.wantAbort {
					assert.False(tc.creator.createCalled)
				} else {
					assert.NoError(err)
					assert.True(tc.creator.createCalled)
					assert.Equal(tc.creator.id.AzureOutput, validIAMIDFile.AzureOutput)
				}
			}
		})
	}
}
