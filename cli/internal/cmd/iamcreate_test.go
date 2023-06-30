/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIDFile(t *testing.T) {
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: base64.RawStdEncoding.EncodeToString([]byte(`{"private_key_id":"not_a_secret"}`)),
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

func TestIAMCreateAWS(t *testing.T) {
	defaultFs := createFSWithConfig(*createConfig(cloudprovider.AWS))
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
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
		setupFs             func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs
		creator             *stubIAMCreator
		provider            cloudprovider.Provider
		zoneFlag            string
		prefixFlag          string
		yesFlag             bool
		updateConfigFlag    bool
		configFlag          string
		existingConfigFiles []string
		existingDirs        []string
		stdin               string
		wantAbort           bool
		wantErr             bool
	}{
		"iam create aws": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			yesFlag:             true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"iam create aws fails when --zone has no availability zone": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-1",
			prefixFlag:          "test",
			yesFlag:             true,
			existingConfigFiles: []string{constants.ConfigFilename},
			wantErr:             true,
		},
		"iam create aws --update-config": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			yesFlag:             true,
			configFlag:          constants.ConfigFilename,
			updateConfigFlag:    true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"iam create aws --update-config fails when --zone is different from zone in config": {
			setupFs: createFSWithConfig(func() config.Config {
				cfg := createConfig(cloudprovider.AWS)
				cfg.Provider.AWS.Zone = "eu-central-1a"
				cfg.Provider.AWS.Region = "eu-central-1"
				return *cfg
			}()),
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-1a",
			prefixFlag:          "test",
			yesFlag:             true,
			existingConfigFiles: []string{constants.ConfigFilename},
			updateConfigFlag:    true,
			wantErr:             true,
		},
		"iam create aws --update-config fails when config has different provider": {
			setupFs:             createFSWithConfig(*createConfig(cloudprovider.GCP)),
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-1a",
			prefixFlag:          "test",
			yesFlag:             true,
			existingConfigFiles: []string{constants.ConfigFilename},
			updateConfigFlag:    true,
			wantErr:             true,
		},
		"iam create aws no config": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
		},
		"iam create aws --update-config with --config": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			yesFlag:             true,
			updateConfigFlag:    true,
			configFlag:          "custom-config.yaml",
			existingConfigFiles: []string{"custom-config.yaml"},
		},
		"iam create aws --update-config --config path doesn't exist": {
			setupFs:          defaultFs,
			creator:          &stubIAMCreator{id: validIAMIDFile},
			provider:         cloudprovider.AWS,
			zoneFlag:         "us-east-2a",
			prefixFlag:       "test",
			yesFlag:          true,
			updateConfigFlag: true,
			wantErr:          true,
			configFlag:       constants.ConfigFilename,
		},
		"iam create aws existing terraform dir": {
			setupFs:      defaultFs,
			creator:      &stubIAMCreator{id: validIAMIDFile},
			provider:     cloudprovider.AWS,
			zoneFlag:     "us-east-2a",
			prefixFlag:   "test",
			yesFlag:      true,
			wantErr:      true,
			existingDirs: []string{constants.TerraformIAMWorkingDir},
		},
		"interactive": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "yes\n",
		},
		"interactive update config": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			stdin:               "yes\n",
			configFlag:          constants.ConfigFilename,
			updateConfigFlag:    true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"interactive abort": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "no\n",
			wantAbort:  true,
		},
		"interactive update config abort": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			provider:            cloudprovider.AWS,
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			stdin:               "no\n",
			updateConfigFlag:    true,
			configFlag:          constants.ConfigFilename,
			wantAbort:           true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"invalid zone": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-west",
			prefixFlag: "test",
			yesFlag:    true,
			wantErr:    true,
		},
		"unwritable fs": {
			setupFs:          readOnlyFs,
			creator:          &stubIAMCreator{id: validIAMIDFile},
			provider:         cloudprovider.AWS,
			zoneFlag:         "us-east-2a",
			prefixFlag:       "test",
			yesFlag:          true,
			updateConfigFlag: true,
			wantErr:          true,
			configFlag:       constants.ConfigFilename,
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

			// register persistent flags manually
			cmd.Flags().String("config", constants.ConfigFilename, "")
			cmd.Flags().Bool("update-config", false, "")
			cmd.Flags().Bool("yes", false, "")
			cmd.Flags().String("name", "constell", "")
			cmd.Flags().String("tf-log", "NONE", "")

			if tc.zoneFlag != "" {
				require.NoError(cmd.Flags().Set("zone", tc.zoneFlag))
			}
			if tc.prefixFlag != "" {
				require.NoError(cmd.Flags().Set("prefix", tc.prefixFlag))
			}
			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}
			if tc.updateConfigFlag {
				require.NoError(cmd.Flags().Set("update-config", "true"))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingConfigFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:             cmd,
				log:             logger.NewTest(t),
				spinner:         &nopSpinner{},
				creator:         tc.creator,
				fileHandler:     fileHandler,
				iamConfig:       &cloudcmd.IAMConfigOptions{},
				provider:        tc.provider,
				providerCreator: &awsIAMCreator{},
			}
			err := iamCreator.create(cmd.Context())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			if tc.wantAbort {
				assert.False(tc.creator.createCalled)
				return
			}

			if tc.updateConfigFlag {
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
		})
	}
}

func TestIAMCreateAzure(t *testing.T) {
	defaultFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingConfigFiles {
			require.NoError(fileHandler.WriteYAML(f, createConfig(cloudprovider.Azure), file.OptNone))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
		return fs
	}
	validIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: iamid.AzureFile{
			SubscriptionID: "test_subscription_id",
			TenantID:       "test_tenant_id",
			UAMIID:         "test_uami_id",
		},
	}

	testCases := map[string]struct {
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		regionFlag           string
		servicePrincipalFlag string
		resourceGroupFlag    string
		yesFlag              bool
		updateConfigFlag     bool
		configFlag           string
		existingConfigFiles  []string
		existingDirs         []string
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
		"iam create azure with existing config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create azure --update-config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			updateConfigFlag:     true,
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create azure --update-config with --config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			updateConfigFlag:     true,
			configFlag:           "custom-config.yaml",
			yesFlag:              true,
			existingConfigFiles:  []string{"custom-config.yaml"},
		},
		"iam create azure --update-config custom --config path doesn't exist": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			updateConfigFlag:     true,
			yesFlag:              true,
			wantErr:              true,
			configFlag:           "custom-config.yaml",
		},
		"iam create azur --update-config --config path doesn't exists": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			updateConfigFlag:     true,
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
			wantErr:              true,
		},
		"iam create azure existing terraform dir": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
			wantErr:              true,
			existingDirs:         []string{constants.TerraformIAMWorkingDir},
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
		"interactive update config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "yes\n",
			updateConfigFlag:     true,
			configFlag:           constants.ConfigFilename,
			existingConfigFiles:  []string{constants.ConfigFilename},
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
		"interactive update config abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "no\n",
			updateConfigFlag:     true,
			wantAbort:            true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"unwritable fs": {
			setupFs:              readOnlyFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
			updateConfigFlag:     true,
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

			// register persistent flags manually
			cmd.Flags().String("config", constants.ConfigFilename, "")
			cmd.Flags().Bool("update-config", false, "")
			cmd.Flags().Bool("yes", false, "")
			cmd.Flags().String("name", "constell", "")
			cmd.Flags().String("tf-log", "NONE", "")

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
			if tc.updateConfigFlag {
				require.NoError(cmd.Flags().Set("update-config", "true"))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingConfigFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:             cmd,
				log:             logger.NewTest(t),
				spinner:         &nopSpinner{},
				creator:         tc.creator,
				fileHandler:     fileHandler,
				iamConfig:       &cloudcmd.IAMConfigOptions{},
				provider:        tc.provider,
				providerCreator: &azureIAMCreator{},
			}
			err := iamCreator.create(cmd.Context())

			if tc.wantErr {
				assert.Error(err)
				return
			}

			if tc.wantAbort {
				assert.False(tc.creator.createCalled)
				return
			}

			if tc.updateConfigFlag {
				readConfig := &config.Config{}
				readErr := fileHandler.ReadYAML(tc.configFlag, readConfig)
				require.NoError(readErr)
				assert.Equal(tc.creator.id.AzureOutput.SubscriptionID, readConfig.Provider.Azure.SubscriptionID)
				assert.Equal(tc.creator.id.AzureOutput.TenantID, readConfig.Provider.Azure.TenantID)
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

func TestIAMCreateGCP(t *testing.T) {
	defaultFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingConfigFiles {
			require.NoError(fileHandler.WriteYAML(f, createConfig(cloudprovider.GCP), file.OptNone))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
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
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		zoneFlag             string
		serviceAccountIDFlag string
		projectIDFlag        string
		yesFlag              bool
		updateConfigFlag     bool
		configFlag           string
		existingConfigFiles  []string
		existingDirs         []string
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
		"iam create gcp with existing config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create gcp --update-config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			updateConfigFlag:     true,
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create gcp --update-config with --config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			updateConfigFlag:     true,
			configFlag:           "custom-config.yaml",
			yesFlag:              true,
			existingConfigFiles:  []string{"custom-config.yaml"},
		},
		"iam create gcp --update-config --config path doesn't exists": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			updateConfigFlag:     true,
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
			wantErr:              true,
		},
		"iam create gcp --update-config wrong --config path": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			updateConfigFlag:     true,
			configFlag:           "custom-config.yaml",
			yesFlag:              true,
			wantErr:              true,
		},
		"iam create gcp existing terraform dir": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",

			existingDirs: []string{constants.TerraformIAMWorkingDir},
			yesFlag:      true,
			wantErr:      true,
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
		"interactive update config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "yes\n",
			configFlag:           constants.ConfigFilename,
			updateConfigFlag:     true,
			existingConfigFiles:  []string{constants.ConfigFilename},
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
		"interactive abort update config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "no\n",
			wantAbort:            true,
			configFlag:           constants.ConfigFilename,
			updateConfigFlag:     true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"unwritable fs": {
			setupFs:              readOnlyFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			updateConfigFlag:     true,
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

			// register persistent flags manually
			cmd.Flags().String("config", constants.ConfigFilename, "")
			cmd.Flags().Bool("update-config", false, "")
			cmd.Flags().Bool("yes", false, "")
			cmd.Flags().String("name", "constell", "")
			cmd.Flags().String("tf-log", "NONE", "")

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
			if tc.updateConfigFlag {
				require.NoError(cmd.Flags().Set("update-config", "true"))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingConfigFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:             cmd,
				log:             logger.NewTest(t),
				spinner:         &nopSpinner{},
				creator:         tc.creator,
				fileHandler:     fileHandler,
				iamConfig:       &cloudcmd.IAMConfigOptions{},
				provider:        tc.provider,
				providerCreator: &gcpIAMCreator{},
			}
			err := iamCreator.create(cmd.Context())

			if tc.wantErr {
				assert.Error(err)
				return
			}

			if tc.wantAbort {
				assert.False(tc.creator.createCalled)
				return
			}

			if tc.updateConfigFlag {
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

func TestValidateConfigWithFlagCompatibility(t *testing.T) {
	testCases := map[string]struct {
		iamProvider cloudprovider.Provider
		cfg         config.Config
		flags       iamFlags
		wantErr     bool
	}{
		"AWS valid when cfg.zone == flag.zone": {
			iamProvider: cloudprovider.AWS,
			cfg: func() config.Config {
				cfg := createConfig(cloudprovider.AWS)
				cfg.Provider.AWS.Zone = "europe-west-1a"
				return *cfg
			}(),
			flags: iamFlags{
				aws: awsFlags{
					zone: "europe-west-1a",
				},
			},
		},
		"AWS valid when cfg.zone not set": {
			iamProvider: cloudprovider.AWS,
			cfg:         *createConfig(cloudprovider.AWS),
			flags: iamFlags{
				aws: awsFlags{
					zone: "europe-west-1a",
				},
			},
		},
		"GCP invalid when cfg.zone != flag.zone": {
			iamProvider: cloudprovider.GCP,
			cfg: func() config.Config {
				cfg := createConfig(cloudprovider.GCP)
				cfg.Provider.GCP.Zone = "europe-west-1a"
				return *cfg
			}(),
			flags: iamFlags{
				aws: awsFlags{
					zone: "us-west-1a",
				},
			},
			wantErr: true,
		},
		"Azure invalid when cfg.zone != flag.zone": {
			iamProvider: cloudprovider.GCP,
			cfg: func() config.Config {
				cfg := createConfig(cloudprovider.Azure)
				cfg.Provider.Azure.Location = "europe-west-1a"
				return *cfg
			}(),
			flags: iamFlags{
				aws: awsFlags{
					zone: "us-west-1a",
				},
			},
			wantErr: true,
		},
		"GCP invalid when cfg.provider different from iam provider": {
			iamProvider: cloudprovider.GCP,
			cfg:         *createConfig(cloudprovider.AWS),
			wantErr:     true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := ValidateConfigWithFlagCompatibility(tc.iamProvider, tc.cfg, tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func createFSWithConfig(cfg config.Config) func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
	return func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingConfigFiles {
			require.NoError(fileHandler.WriteYAML(f, cfg, file.OptNone))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
}
