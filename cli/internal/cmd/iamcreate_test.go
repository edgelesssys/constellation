/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
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
	defaultFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingFiles {
			require.NoError(fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs {
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
		setupFs            func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs
		creator            *stubIAMCreator
		provider           cloudprovider.Provider
		zoneFlag           string
		prefixFlag         string
		yesFlag            bool
		generateConfigFlag bool
		k8sVersionFlag     string
		configFlag         string
		existingFiles      []string
		existingDirs       []string
		stdin              string
		wantAbort          bool
		wantErr            bool
	}{
		"iam create aws": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-east-2a",
			prefixFlag: "test",
			yesFlag:    true,
		},
		"iam create aws generate config": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			yesFlag:            true,
			configFlag:         constants.ConfigFilename,
			generateConfigFlag: true,
		},
		"iam create aws generate config custom path": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			yesFlag:            true,
			generateConfigFlag: true,
			configFlag:         "custom-config.yaml",
		},
		"iam create aws generate config path already exists": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			yesFlag:            true,
			generateConfigFlag: true,
			wantErr:            true,
			configFlag:         constants.ConfigFilename,
			existingFiles:      []string{constants.ConfigFilename},
		},
		"iam create aws generate config custom path already exists": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			yesFlag:            true,
			generateConfigFlag: true,
			wantErr:            true,
			configFlag:         "custom-config.yaml",
			existingFiles:      []string{"custom-config.yaml"},
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
		"interactive generate config": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			stdin:              "yes\n",
			configFlag:         constants.ConfigFilename,
			generateConfigFlag: true,
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
		"interactive generate config abort": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			stdin:              "no\n",
			generateConfigFlag: true,
			configFlag:         constants.ConfigFilename,
			wantAbort:          true,
		},
		"invalid zone": {
			setupFs:    defaultFs,
			creator:    &stubIAMCreator{id: validIAMIDFile},
			provider:   cloudprovider.AWS,
			zoneFlag:   "us-west-5b",
			prefixFlag: "test",
			yesFlag:    true,
			wantErr:    true,
		},
		"unwritable fs": {
			setupFs:            readOnlyFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			yesFlag:            true,
			generateConfigFlag: true,
			wantErr:            true,
			configFlag:         constants.ConfigFilename,
		},
		"iam create azure without generate config and invalid kubernetes version": {
			setupFs:        defaultFs,
			creator:        &stubIAMCreator{id: validIAMIDFile},
			provider:       cloudprovider.AWS,
			zoneFlag:       "us-east-2a",
			prefixFlag:     "test",
			k8sVersionFlag: "1.11.1", // supposed to be ignored without generateConfigFlag
			yesFlag:        true,
		},
		"iam create azure generate config with valid kubernetes version": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			generateConfigFlag: true,
			k8sVersionFlag:     semver.MajorMinor(string(versions.Default)),
			configFlag:         constants.ConfigFilename,
			yesFlag:            true,
		},
		"iam create azure generate config with invalid kubernetes version": {
			setupFs:            defaultFs,
			creator:            &stubIAMCreator{id: validIAMIDFile},
			provider:           cloudprovider.AWS,
			zoneFlag:           "us-east-2a",
			prefixFlag:         "test",
			generateConfigFlag: true,
			k8sVersionFlag:     "1.22.1",
			configFlag:         constants.ConfigFilename,
			yesFlag:            true,
			wantErr:            true,
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
			cmd.Flags().Bool("generate-config", false, "")
			cmd.Flags().String("kubernetes", semver.MajorMinor(config.Default().KubernetesVersion), "")
			cmd.Flags().Bool("yes", false, "")
			cmd.Flags().String("name", "constell", "")

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
			if tc.k8sVersionFlag != "" {
				require.NoError(cmd.Flags().Set("kubernetes", tc.k8sVersionFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:             cmd,
				log:             logger.NewTest(t),
				spinner:         &nopSpinner{},
				creator:         tc.creator,
				fileHandler:     fileHandler,
				iamConfig:       &cloudcmd.IAMConfig{},
				provider:        tc.provider,
				providerCreator: &awsIAMCreator{},
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
		})
	}
}

func TestIAMCreateAzure(t *testing.T) {
	defaultFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingFiles {
			require.NoError(fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs {
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
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		regionFlag           string
		servicePrincipalFlag string
		resourceGroupFlag    string
		yesFlag              bool
		generateConfigFlag   bool
		k8sVersionFlag       string
		configFlag           string
		existingFiles        []string
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
		"iam create azure without generate config and invalid kubernetes version": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			k8sVersionFlag:       "1.11.1", // supposed to be ignored without generateConfigFlag
			yesFlag:              true,
		},
		"iam create azure generate config with valid kubernetes version": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			generateConfigFlag:   true,
			k8sVersionFlag:       semver.MajorMinor(string(versions.Default)),
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
		},
		"iam create azure generate config with invalid kubernetes version": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.Azure,
			regionFlag:           "westus",
			servicePrincipalFlag: "constell-test-sp",
			resourceGroupFlag:    "constell-test-rg",
			generateConfigFlag:   true,
			k8sVersionFlag:       "1.22.1",
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
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

			// register persistent flag manually
			cmd.Flags().String("config", constants.ConfigFilename, "")
			cmd.Flags().Bool("generate-config", false, "")
			cmd.Flags().String("kubernetes", semver.MajorMinor(config.Default().KubernetesVersion), "")
			cmd.Flags().Bool("yes", false, "")
			cmd.Flags().String("name", "constell", "")

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
			if tc.k8sVersionFlag != "" {
				require.NoError(cmd.Flags().Set("kubernetes", tc.k8sVersionFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:             cmd,
				log:             logger.NewTest(t),
				spinner:         &nopSpinner{},
				creator:         tc.creator,
				fileHandler:     fileHandler,
				iamConfig:       &cloudcmd.IAMConfig{},
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

func TestIAMCreateGCP(t *testing.T) {
	defaultFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		for _, f := range existingFiles {
			require.NoError(fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
		}
		for _, d := range existingDirs {
			require.NoError(fs.MkdirAll(d, 0o755))
		}
		return fs
	}
	readOnlyFs := func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs {
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
		setupFs              func(require *require.Assertions, provider cloudprovider.Provider, existingFiles []string, existingDirs []string) afero.Fs
		creator              *stubIAMCreator
		provider             cloudprovider.Provider
		zoneFlag             string
		serviceAccountIDFlag string
		projectIDFlag        string
		yesFlag              bool
		generateConfigFlag   bool
		k8sVersionFlag       string
		configFlag           string
		existingFiles        []string
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
		"iam create gcp without generate config and invalid kubernetes version": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			k8sVersionFlag:       "1.11.1", // supposed to be ignored without generateConfigFlag
			yesFlag:              true,
		},
		"iam create gcp generate config with valid kubernetes version": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			generateConfigFlag:   true,
			k8sVersionFlag:       semver.MajorMinor(string(versions.Default)),
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
		},
		"iam create gcp generate config with invalid kubernetes version": {
			setupFs:              defaultFs,
			creator:              &stubIAMCreator{id: validIAMIDFile},
			provider:             cloudprovider.GCP,
			zoneFlag:             "europe-west1-a",
			serviceAccountIDFlag: "constell-test",
			projectIDFlag:        "constell-1234",
			generateConfigFlag:   true,
			k8sVersionFlag:       "1.22.1",
			configFlag:           constants.ConfigFilename,
			yesFlag:              true,
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
			cmd.Flags().Bool("generate-config", false, "")
			cmd.Flags().String("kubernetes", semver.MajorMinor(config.Default().KubernetesVersion), "")
			cmd.Flags().Bool("yes", false, "")
			cmd.Flags().String("name", "constell", "")

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
			if tc.k8sVersionFlag != "" {
				require.NoError(cmd.Flags().Set("kubernetes", tc.k8sVersionFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider, tc.existingFiles, tc.existingDirs))

			iamCreator := &iamCreator{
				cmd:             cmd,
				log:             logger.NewTest(t),
				spinner:         &nopSpinner{},
				creator:         tc.creator,
				fileHandler:     fileHandler,
				iamConfig:       &cloudcmd.IAMConfig{},
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
