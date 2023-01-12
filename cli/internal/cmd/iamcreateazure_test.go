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
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		regionFlag           string
		servicePrincipalFlag string
		resourceGroupFlag    string
		yesFlag              bool
		generateConfigFlag   bool
		configFlag           string
		existingFiles        []string
		stdin                string
		wantAbort            bool
		wantErr              bool
	}{
		"iam create azure": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
		},
		"iam create azure generate config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			generateConfigFlag:   true,
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
		},
		"iam create azure generate config custom path": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			generateConfigFlag:   true,
			configFlag:           "custom-config.yaml",
			yesFlag:              true,
		},
		"iam create azure generate config custom path already exists": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			generateConfigFlag:   true,
			yesFlag:              true,
			wantErr:              true,
			configFlag:           "custom-config.yaml",
			existingFiles:        []string{"custom-config.yaml"},
		},
		"iam create generate config path already exists": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			generateConfigFlag:   true,
			configFlag:           constants.ConfigFilename,
			existingFiles:        []string{constants.ConfigFilename},
			yesFlag:              true,
			wantErr:              true,
		},
		"interactive": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "yes\n",
		},
		"interactive generate config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "yes\n",
			generateConfigFlag:   true,
			configFlag:           constants.ConfigFilename,
		},
		"interactive abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "no\n",
			wantAbort:            true,
		},
		"interactive generate config abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "no\n",
			generateConfigFlag:   true,
			wantAbort:            true,
		},
		"unwritable fs": {
			setupFs:              readOnlyFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
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

			cmd := newIAMCreateAzureCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("generate-config", false, "")             // register persistent flag manually

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
			if tc.generateConfigFlag {
				require.NoError(cmd.Flags().Set("generate-config", "true"))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}


			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingFiles))

			err := iamCreateAzure(cmd, nopSpinner{}, tc.creator, fileHandler)

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
				assert.Equal(tc.creator.id.AzureOutput.SubscriptionID, readConfig.Provider.Azure.SubscriptionID)
				assert.Equal(tc.creator.id.AzureOutput.TenantID, readConfig.Provider.Azure.TenantID)
				assert.Equal(tc.creator.id.AzureOutput.ApplicationID, readConfig.Provider.Azure.AppClientID)
				assert.Equal(tc.creator.id.AzureOutput.ApplicationClientSecretValue, readConfig.Provider.Azure.ClientSecretValue)
				assert.Equal(tc.creator.id.AzureOutput.UAMIID, readConfig.Provider.Azure.UserAssignedIdentity)
				assert.Equal(tc.regionFlag, readConfig.Provider.Azure.Location)
				assert.Equal(tc.resourceGroupFlag, readConfig.Provider.Azure.ResourceGroup)
			}
			require.NoError(err)
			assert.True(tc.creator.createCalled)
			assert.Equal(tc.creator.id.AzureOutput, validIAMIDFile.AzureOutput)
		})
	}
}
