/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"encoding/base64"
	"log/slog"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
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
	validIAMIDFile := cloudcmd.IAMOutput{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: cloudcmd.GCPIAMOutput{
			ServiceAccountKey: base64.RawStdEncoding.EncodeToString([]byte(`{"private_key_id":"not_a_secret"}`)),
		},
	}
	invalidIAMIDFile := cloudcmd.IAMOutput{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: cloudcmd.GCPIAMOutput{
			ServiceAccountKey: "ey_Jwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // invalid b64
		},
	}
	testCases := map[string]struct {
		idFile           cloudcmd.IAMOutput
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
	validIAMIDFile := cloudcmd.IAMOutput{
		CloudProvider: cloudprovider.AWS,
		AWSOutput: cloudcmd.AWSIAMOutput{
			ControlPlaneInstanceProfile: "test_control_plane_instance_profile",
			WorkerNodeInstanceProfile:   "test_worker_nodes_instance_profile_name",
		},
	}

	testCases := map[string]struct {
		setupFs             func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs
		creator             *stubIAMCreator
		zoneFlag            string
		prefixFlag          string
		yesFlag             bool
		updateConfigFlag    bool
		existingConfigFiles []string
		existingDirs        []string
		stdin               string
		wantAbort           bool
		wantErr             bool
	}{
		"iam create aws": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			yesFlag:             true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"iam create aws --update-config": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			yesFlag:             true,
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
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
		},
		"iam create aws existing terraform dir": {
			setupFs:      defaultFs,
			creator:      &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:     "us-east-2a",
			prefixFlag:   "test",
			yesFlag:      true,
			wantErr:      true,
			existingDirs: []string{constants.TerraformIAMWorkingDir},
		},
		"interactive": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "yes\n",
		},
		"interactive update config": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			stdin:               "yes\n",
			updateConfigFlag:    true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"interactive abort": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			stdin:      "no\n",
			wantAbort:  true,
		},
		"interactive update config abort": {
			setupFs:             defaultFs,
			creator:             &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:            "us-east-2a",
			prefixFlag:          "test",
			stdin:               "no\n",
			updateConfigFlag:    true,
			wantAbort:           true,
			existingConfigFiles: []string{constants.ConfigFilename},
		},
		"unwritable fs": {
			setupFs:          readOnlyFs,
			creator:          &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:         "us-east-2a",
			prefixFlag:       "test",
			yesFlag:          true,
			updateConfigFlag: true,
			wantErr:          true,
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

			fileHandler := file.NewHandler(tc.setupFs(require, cloudprovider.AWS, tc.existingConfigFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:         cmd,
				log:         slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				spinner:     &nopSpinner{},
				creator:     tc.creator,
				fileHandler: fileHandler,
				iamConfig:   &cloudcmd.IAMConfigOptions{},
				provider:    cloudprovider.AWS,
				flags: iamCreateFlags{
					yes:          tc.yesFlag,
					updateConfig: tc.updateConfigFlag,
				},
				providerCreator: &awsIAMCreator{
					flags: awsIAMCreateFlags{
						zone:   tc.zoneFlag,
						prefix: tc.prefixFlag,
					},
				},
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
				readErr := fileHandler.ReadYAML(constants.ConfigFilename, readConfig)
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
	defaultFs := createFSWithConfig(*createConfig(cloudprovider.Azure))
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
		return fs
	}
	validIAMIDFile := cloudcmd.IAMOutput{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: cloudcmd.AzureIAMOutput{
			SubscriptionID: "test_subscription_id",
			TenantID:       "test_tenant_id",
			UAMIID:         "test_uami_id",
		},
	}

	testCases := map[string]struct {
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs
		creator              *stubIAMCreator
		regionFlag           string
		servicePrincipalFlag string
		resourceGroupFlag    string
		yesFlag              bool
		updateConfigFlag     bool
		existingConfigFiles  []string
		existingDirs         []string
		stdin                string
		wantAbort            bool
		wantErr              bool
	}{
		"iam create azure": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
		},
		"iam create azure with existing config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create azure --update-config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			updateConfigFlag:     true,
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create azure existing terraform dir": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
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
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "yes\n",
		},
		"interactive update config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "yes\n",
			updateConfigFlag:     true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"interactive abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			stdin:                "no\n",
			wantAbort:            true,
		},
		"interactive update config abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
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
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			yesFlag:              true,
			updateConfigFlag:     true,
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

			fileHandler := file.NewHandler(tc.setupFs(require, cloudprovider.Azure, tc.existingConfigFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:         cmd,
				log:         slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				spinner:     &nopSpinner{},
				creator:     tc.creator,
				fileHandler: fileHandler,
				iamConfig:   &cloudcmd.IAMConfigOptions{},
				provider:    cloudprovider.Azure,
				flags: iamCreateFlags{
					yes:          tc.yesFlag,
					updateConfig: tc.updateConfigFlag,
				},
				providerCreator: &azureIAMCreator{
					flags: azureIAMCreateFlags{
						region:           tc.regionFlag,
						resourceGroup:    tc.resourceGroupFlag,
						servicePrincipal: tc.servicePrincipalFlag,
					},
				},
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
				readErr := fileHandler.ReadYAML(constants.ConfigFilename, readConfig)
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
	defaultFs := createFSWithConfig(*createConfig(cloudprovider.GCP))
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
		return fs
	}
	validIAMIDFile := cloudcmd.IAMOutput{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: cloudcmd.GCPIAMOutput{
			ServiceAccountKey: "eyJwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // {"private_key_id":"not_a_secret"}
		},
	}
	invalidIAMIDFile := cloudcmd.IAMOutput{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: cloudcmd.GCPIAMOutput{
			ServiceAccountKey: "ey_Jwcml2YXRlX2tleV9pZCI6Im5vdF9hX3NlY3JldCJ9Cg==", // invalid b64
		},
	}

	testCases := map[string]struct {
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingConfigFiles []string, existingDirs []string) afero.Fs
		creator              *stubIAMCreator
		zoneFlag             string
		serviceAccountIDFlag string
		projectIDFlag        string
		yesFlag              bool
		updateConfigFlag     bool
		existingConfigFiles  []string
		existingDirs         []string
		stdin                string
		wantAbort            bool
		wantErr              bool
	}{
		"iam create gcp": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
		},
		"iam create gcp with existing config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create gcp --update-config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			updateConfigFlag:     true,
			yesFlag:              true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"iam create gcp existing terraform dir": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",

			existingDirs: []string{constants.TerraformIAMWorkingDir},
			yesFlag:      true,
			wantErr:      true,
		},
		"iam create gcp invalid b64": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: invalidIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			wantErr:              true,
		},
		"interactive": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "yes\n",
		},
		"interactive update config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "yes\n",
			updateConfigFlag:     true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"interactive abort": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "no\n",
			wantAbort:            true,
		},
		"interactive abort update config": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			stdin:                "no\n",
			wantAbort:            true,
			updateConfigFlag:     true,
			existingConfigFiles:  []string{constants.ConfigFilename},
		},
		"unwritable fs": {
			setupFs:              readOnlyFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			yesFlag:              true,
			updateConfigFlag:     true,
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

			fileHandler := file.NewHandler(tc.setupFs(require, cloudprovider.GCP, tc.existingConfigFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:         cmd,
				log:         slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				spinner:     &nopSpinner{},
				creator:     tc.creator,
				fileHandler: fileHandler,
				iamConfig:   &cloudcmd.IAMConfigOptions{},
				provider:    cloudprovider.GCP,
				flags: iamCreateFlags{
					yes:          tc.yesFlag,
					updateConfig: tc.updateConfigFlag,
				},
				providerCreator: &gcpIAMCreator{
					flags: gcpIAMCreateFlags{
						zone:             tc.zoneFlag,
						serviceAccountID: tc.serviceAccountIDFlag,
						projectID:        tc.projectIDFlag,
					},
				},
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
				readErr := fileHandler.ReadYAML(constants.ConfigFilename, readConfig)
				require.NoError(readErr)
				assert.Equal(constants.GCPServiceAccountKeyFilename, readConfig.Provider.GCP.ServiceAccountKeyPath)
			}
			require.NoError(err)
			assert.True(tc.creator.createCalled)
			assert.Equal(tc.creator.id.GCPOutput, validIAMIDFile.GCPOutput)
			readServiceAccountKey := &map[string]string{}
			readErr := fileHandler.ReadJSON(constants.GCPServiceAccountKeyFilename, readServiceAccountKey)
			require.NoError(readErr)
			assert.Equal("not_a_secret", (*readServiceAccountKey)["private_key_id"])
		})
	}
}

func TestValidateConfigWithFlagCompatibility(t *testing.T) {
	testCases := map[string]struct {
		iamProvider cloudprovider.Provider
		cfg         config.Config
		zone        string
		wantErr     bool
	}{
		"AWS valid when cfg.zone == flag.zone": {
			iamProvider: cloudprovider.AWS,
			cfg: func() config.Config {
				cfg := createConfig(cloudprovider.AWS)
				cfg.Provider.AWS.Zone = "europe-west-1a"
				return *cfg
			}(),
			zone: "europe-west-1a",
		},
		"AWS valid when cfg.zone not set": {
			iamProvider: cloudprovider.AWS,
			cfg:         *createConfig(cloudprovider.AWS),
			zone:        "europe-west-1a",
		},
		"GCP invalid when cfg.zone != flag.zone": {
			iamProvider: cloudprovider.GCP,
			cfg: func() config.Config {
				cfg := createConfig(cloudprovider.GCP)
				cfg.Provider.GCP.Zone = "europe-west-1a"
				return *cfg
			}(),
			zone:    "us-west-1a",
			wantErr: true,
		},
		"Azure invalid when cfg.zone != flag.zone": {
			iamProvider: cloudprovider.GCP,
			cfg: func() config.Config {
				cfg := createConfig(cloudprovider.Azure)
				cfg.Provider.Azure.Location = "europe-west-1a"
				return *cfg
			}(),
			zone:    "us-west-1a",
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
			err := validateConfigWithFlagCompatibility(tc.iamProvider, tc.cfg, tc.zone)
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
			require.NoError(fileHandler.WriteYAML(f, cfg, file.OptMkdirAll))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
}
