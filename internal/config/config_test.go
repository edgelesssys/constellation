/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"gopkg.in/yaml.v3"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	def := DefaultWithPanic()
	assert.NotNil(def)
}

func TestSettingLatestAsVersion(t *testing.T) {
	testCases := map[string]struct {
		config     map[string]interface{}
		configName string
		wantResult *Config
		wantErr    bool
	}{
		"mix of latest and uint as version value": {
			config: func() map[string]interface{} {
				conf := DefaultWithPanic()
				// modify versions as string
				m := getConfigAsMap(conf, t)
				m["attestation"].(map[string]interface{})["azureSEVSNP"].(map[string]interface{})["microcodeVersion"] = "latest"
				m["attestation"].(map[string]interface{})["azureSEVSNP"].(map[string]interface{})["teeVersion"] = "latest"
				m["attestation"].(map[string]interface{})["azureSEVSNP"].(map[string]interface{})["snpVersion"] = "latest"
				m["attestation"].(map[string]interface{})["azureSEVSNP"].(map[string]interface{})["bootloaderVersion"] = 1
				return m
			}(),

			configName: constants.ConfigFilename,
			wantResult: func() *Config {
				conf := DefaultWithPanic()
				conf.Attestation.AzureSEVSNP.BootloaderVersion = 1
				return conf
			}(),
		},
		"refuse invalid version value": {
			config: func() map[string]interface{} {
				conf := DefaultWithPanic()
				m := getConfigAsMap(conf, t)
				m["attestation"].(map[string]interface{})["azureSEVSNP"].(map[string]interface{})["microcodeVersion"] = "1a"
				return m
			}(),
			configName: constants.ConfigFilename,
			wantErr:    true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			if tc.config != nil {
				require.NoError(fileHandler.WriteYAML(tc.configName, tc.config, file.OptNone))
			}
			result, err := fromFile(fileHandler, tc.configName)

			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(tc.wantResult, result)
			}
		})
	}
}

// getConfigAsMap returns a map of the attestation section of the config.
func getConfigAsMap(conf *Config, t *testing.T) map[string]interface{} {
	d, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	err = yaml.Unmarshal(d, &m)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestFromFile(t *testing.T) {
	testCases := map[string]struct {
		config     *Config
		configName string
		wantResult *Config
		wantErr    bool
	}{
		"default config from default file": {
			config:     DefaultWithPanic(),
			configName: constants.ConfigFilename,
			wantResult: DefaultWithPanic(),
		},
		"default config from different path": {
			config:     DefaultWithPanic(),
			configName: "other-config.yaml",
			wantResult: DefaultWithPanic(),
		},
		"default config when path empty": {
			config:     nil,
			configName: "",
			wantErr:    true,
		},
		"err when path not exist": {
			config:     nil,
			configName: "wrong-name.yaml",
			wantErr:    true,
		},
		"custom config from default file": {
			config: &Config{
				Version: Version3,
			},
			configName: constants.ConfigFilename,
			wantResult: &Config{
				Version: Version3,
			},
		},
		"modify default config": {
			config: func() *Config {
				conf := DefaultWithPanic()
				conf.Provider.GCP.Region = "eu-north1"
				conf.Provider.GCP.Zone = "eu-north1-a"
				return conf
			}(),
			configName: constants.ConfigFilename,
			wantResult: func() *Config {
				conf := DefaultWithPanic()
				conf.Provider.GCP.Region = "eu-north1"
				conf.Provider.GCP.Zone = "eu-north1-a"
				return conf
			}(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			if tc.config != nil {
				require.NoError(fileHandler.WriteYAML(tc.configName, tc.config, file.OptNone))
			}

			result, err := fromFile(fileHandler, tc.configName)

			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(tc.wantResult, result)
			}
		})
	}
}

func TestNewWithDefaultOptions(t *testing.T) {
	testCases := map[string]struct {
		confToWrite           *Config
		envToSet              map[string]string
		wantErr               bool
		wantClientSecretValue string
	}{
		"set env works": {
			confToWrite: func() *Config { // valid config with all, but clientSecretValue
				c := DefaultWithPanic()
				c.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				c.Image = "v" + constants.VersionInfo()
				c.Provider.Azure.SubscriptionID = "f4278079-288c-4766-a98c-ab9d5dba01a5"
				c.Provider.Azure.TenantID = "d4ff9d63-6d6d-4042-8f6a-21e804add5aa"
				c.Provider.Azure.Location = "westus"
				c.Provider.Azure.ResourceGroup = "test"
				c.Provider.Azure.UserAssignedIdentity = "/subscriptions/8b8bd01f-efd9-4113-9bd1-c82137c32da7/resourcegroups/constellation-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-identity"
				c.Provider.Azure.AppClientID = "3ea4bdc1-1cc1-4237-ae78-0831eff3491e"
				c.Attestation.AzureSEVSNP.Measurements = measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
				}
				return c
			}(),
			envToSet: map[string]string{
				constants.EnvVarAzureClientSecretValue: "some-secret",
			},
			wantClientSecretValue: "some-secret",
		},
		"set env overwrites": {
			confToWrite: func() *Config {
				c := DefaultWithPanic()
				c.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				c.Image = "v" + constants.VersionInfo()
				c.Provider.Azure.SubscriptionID = "f4278079-288c-4766-a98c-ab9d5dba01a5"
				c.Provider.Azure.TenantID = "d4ff9d63-6d6d-4042-8f6a-21e804add5aa"
				c.Provider.Azure.Location = "westus"
				c.Provider.Azure.ResourceGroup = "test"
				c.Provider.Azure.ClientSecretValue = "other-value" // < Note secret set in config, as well.
				c.Provider.Azure.UserAssignedIdentity = "/subscriptions/8b8bd01f-efd9-4113-9bd1-c82137c32da7/resourcegroups/constellation-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-identity"
				c.Provider.Azure.AppClientID = "3ea4bdc1-1cc1-4237-ae78-0831eff3491e"
				c.Attestation.AzureSEVSNP.Measurements = measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
				}
				return c
			}(),
			envToSet: map[string]string{
				constants.EnvVarAzureClientSecretValue: "some-secret",
			},
			wantClientSecretValue: "some-secret",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Setup
			fileHandler := file.NewHandler(afero.NewMemMapFs())
			err := fileHandler.WriteYAML(constants.ConfigFilename, tc.confToWrite)
			require.NoError(err)
			for envKey, envValue := range tc.envToSet {
				t.Setenv(envKey, envValue)
			}

			// Test
			c, err := New(fileHandler, constants.ConfigFilename, false)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			var validationErr *ValidationError
			if errors.As(err, &validationErr) {
				t.Log(validationErr.LongMessage())
			}
			assert.Equal(c.Provider.Azure.ClientSecretValue, tc.wantClientSecretValue)
		})
	}
}

func TestValidate(t *testing.T) {
	const defaultErrCount = 34 // expect this number of error messages by default because user-specific values are not set and multiple providers are defined by default
	const azErrCount = 9
	const gcpErrCount = 6

	testCases := map[string]struct {
		cnf          *Config
		wantErr      bool
		wantErrCount int
	}{
		"default config is not valid": {
			cnf:          DefaultWithPanic(),
			wantErr:      true,
			wantErrCount: defaultErrCount,
		},
		"v0 is one error": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.Version = "v0"
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount + 1,
		},
		"v0 and negative state disk are two errors": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.Version = "v0"
				cnf.StateDiskSizeGB = -1
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount + 2,
		},
		"default Azure config is not valid": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: azErrCount,
		},
		"Azure config with all required fields is valid": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				cnf.Image = "v" + constants.VersionInfo()
				az := cnf.Provider.Azure
				az.SubscriptionID = "01234567-0123-0123-0123-0123456789ab"
				az.TenantID = "01234567-0123-0123-0123-0123456789ab"
				az.Location = "test-location"
				az.UserAssignedIdentity = "test-identity"
				az.ResourceGroup = "test-resource-group"
				az.AppClientID = "01234567-0123-0123-0123-0123456789ab"
				az.ClientSecretValue = "test-client-secret"
				cnf.Provider = ProviderConfig{}
				cnf.Provider.Azure = az
				cnf.Attestation.AzureSEVSNP.Measurements = measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
				}
				return cnf
			}(),
		},
		"default GCP config is not valid": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.GCP)
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: gcpErrCount,
		},
		"GCP config with all required fields is valid": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.GCP)
				cnf.Image = "v" + constants.VersionInfo()
				gcp := cnf.Provider.GCP
				gcp.Region = "test-region"
				gcp.Project = "test-project"
				gcp.Zone = "test-zone"
				gcp.ServiceAccountKeyPath = "test-key-path"
				cnf.Provider = ProviderConfig{}
				cnf.Provider.GCP = gcp
				cnf.Attestation.GCPSEVES.Measurements = measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
				}
				return cnf
			}(),
		},
		// TODO: v2.7: remove this test as it should start breaking after v2.6 is released.
		"k8s vMAJOR.MINOR is valid in v2.7": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.KubernetesVersion = "v1.25"
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount,
		},
		// TODO: v2.7: remove this test as it should start breaking after v2.6 is released.
		"k8s MAJOR.MINOR is valid in v2.7": {
			cnf: func() *Config {
				cnf := DefaultWithPanic()
				cnf.KubernetesVersion = "1.25"
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := tc.cnf.Validate(false)

			if tc.wantErr {
				assert.Error(err)
				var valErr *ValidationError
				require.ErrorAs(err, &valErr)
				assert.Equal(tc.wantErrCount, valErr.messagesCount())
				return
			}
			assert.NoError(err)
		})
	}
}

func TestHasProvider(t *testing.T) {
	assert := assert.New(t)
	assert.False((&Config{}).HasProvider(cloudprovider.Unknown))
	assert.False((&Config{}).HasProvider(cloudprovider.Azure))
	assert.False((&Config{}).HasProvider(cloudprovider.GCP))
	assert.False((&Config{}).HasProvider(cloudprovider.QEMU))
	assert.False(DefaultWithPanic().HasProvider(cloudprovider.Unknown))
	assert.True(DefaultWithPanic().HasProvider(cloudprovider.Azure))
	assert.True(DefaultWithPanic().HasProvider(cloudprovider.GCP))
	cnfWithAzure := Config{Provider: ProviderConfig{Azure: &AzureConfig{}}}
	assert.False(cnfWithAzure.HasProvider(cloudprovider.Unknown))
	assert.True(cnfWithAzure.HasProvider(cloudprovider.Azure))
	assert.False(cnfWithAzure.HasProvider(cloudprovider.GCP))
}

func TestConfigRemoveProviderExcept(t *testing.T) {
	testCases := map[string]struct {
		removeExcept cloudprovider.Provider
		wantAWS      *AWSConfig
		wantAzure    *AzureConfig
		wantGCP      *GCPConfig
		wantQEMU     *QEMUConfig
	}{
		"except aws": {
			removeExcept: cloudprovider.AWS,
			wantAWS:      DefaultWithPanic().Provider.AWS,
		},
		"except azure": {
			removeExcept: cloudprovider.Azure,
			wantAzure:    DefaultWithPanic().Provider.Azure,
		},
		"except gcp": {
			removeExcept: cloudprovider.GCP,
			wantGCP:      DefaultWithPanic().Provider.GCP,
		},
		"except qemu": {
			removeExcept: cloudprovider.QEMU,
			wantQEMU:     DefaultWithPanic().Provider.QEMU,
		},
		"unknown provider": {
			removeExcept: cloudprovider.Unknown,
			wantAWS:      DefaultWithPanic().Provider.AWS,
			wantAzure:    DefaultWithPanic().Provider.Azure,
			wantGCP:      DefaultWithPanic().Provider.GCP,
			wantQEMU:     DefaultWithPanic().Provider.QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			conf := DefaultWithPanic()
			conf.RemoveProviderAndAttestationExcept(tc.removeExcept)

			assert.Equal(tc.wantAWS, conf.Provider.AWS)
			assert.Equal(tc.wantAzure, conf.Provider.Azure)
			assert.Equal(tc.wantGCP, conf.Provider.GCP)
			assert.Equal(tc.wantQEMU, conf.Provider.QEMU)
		})
	}
}

func TestConfigGeneratedDocsFresh(t *testing.T) {
	assert := assert.New(t)
	updateMsg := "remember to re-generate config docs! 🔨"

	assert.Len(ConfigDoc.Fields, reflect.ValueOf(Config{}).NumField(), updateMsg)
	assert.Len(ProviderConfigDoc.Fields, reflect.ValueOf(ProviderConfig{}).NumField(), updateMsg)
	assert.Len(AWSConfigDoc.Fields, reflect.ValueOf(AWSConfig{}).NumField(), updateMsg)
	assert.Len(AzureConfigDoc.Fields, reflect.ValueOf(AzureConfig{}).NumField(), updateMsg)
	assert.Len(GCPConfigDoc.Fields, reflect.ValueOf(GCPConfig{}).NumField(), updateMsg)
	assert.Len(QEMUConfigDoc.Fields, reflect.ValueOf(QEMUConfig{}).NumField(), updateMsg)
}

func TestConfig_UpdateMeasurements(t *testing.T) {
	assert := assert.New(t)
	newMeasurements := measurements.M{
		1: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
		2: measurements.WithAllBytes(0x01, measurements.Enforce, measurements.PCRMeasurementLength),
		3: measurements.WithAllBytes(0x02, measurements.Enforce, measurements.PCRMeasurementLength),
	}

	{ // AWS
		conf := DefaultWithPanic()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
		for k := range conf.Attestation.AWSNitroTPM.Measurements {
			delete(conf.Attestation.AWSNitroTPM.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.AWSNitroTPM.Measurements)
	}
	{ // Azure
		conf := DefaultWithPanic()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
		for k := range conf.Attestation.AzureSEVSNP.Measurements {
			delete(conf.Attestation.AzureSEVSNP.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.AzureSEVSNP.Measurements)
	}
	{ // GCP
		conf := DefaultWithPanic()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.GCP)
		for k := range conf.Attestation.GCPSEVES.Measurements {
			delete(conf.Attestation.GCPSEVES.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.GCPSEVES.Measurements)
	}
	{ // QEMU
		conf := DefaultWithPanic()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.QEMU)
		for k := range conf.Attestation.QEMUVTPM.Measurements {
			delete(conf.Attestation.QEMUVTPM.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.QEMUVTPM.Measurements)
	}
}

func TestConfig_IsReleaseImage(t *testing.T) {
	testCases := map[string]struct {
		conf *Config
		want bool
	}{
		"release image v0.0.0": {
			conf: func() *Config {
				conf := DefaultWithPanic()
				conf.Image = "v0.0.0"
				return conf
			}(),
			want: true,
		},
		"branch image": {
			conf: func() *Config {
				conf := DefaultWithPanic()
				conf.Image = "feat-x-vX.Y.Z-pre.0.yyyymmddhhmmss-abcdefabcdef"
				return conf
			}(),
			want: false,
		},
		"debug image": {
			conf: func() *Config {
				conf := DefaultWithPanic()
				conf.Image = "debug-vX.Y.Z-pre.0.yyyymmddhhmmss-abcdefabcdef"
				return conf
			}(),
			want: false,
		},
		"empty config": {
			conf: &Config{},
			want: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.want, tc.conf.IsReleaseImage())
		})
	}
}

func TestValidInstanceTypeForProvider(t *testing.T) {
	testCases := map[string]struct {
		provider       cloudprovider.Provider
		instanceTypes  []string
		nonCVMsAllowed bool
		expectedResult bool
	}{
		"empty all": {
			provider:       cloudprovider.Unknown,
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty aws": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty azure only CVMs": {
			provider:       cloudprovider.Azure,
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty azure with non-CVMs": {
			provider:       cloudprovider.Azure,
			instanceTypes:  []string{},
			nonCVMsAllowed: true,
			expectedResult: false,
		},
		"empty gcp": {
			provider:       cloudprovider.GCP,
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"azure only CVMs": {
			provider:       cloudprovider.Azure,
			instanceTypes:  instancetypes.AzureCVMInstanceTypes,
			expectedResult: true,
		},
		"azure CVMs but CVMs disabled": {
			provider:       cloudprovider.Azure,
			instanceTypes:  instancetypes.AzureCVMInstanceTypes,
			nonCVMsAllowed: true,
			expectedResult: false,
		},
		"azure trusted launch VMs with CVMs enabled": {
			provider:       cloudprovider.Azure,
			instanceTypes:  instancetypes.AzureTrustedLaunchInstanceTypes,
			expectedResult: false,
		},
		"azure trusted launch VMs with CVMs disabled": {
			provider:       cloudprovider.Azure,
			instanceTypes:  instancetypes.AzureTrustedLaunchInstanceTypes,
			nonCVMsAllowed: true,
			expectedResult: true,
		},
		"gcp": {
			provider:       cloudprovider.GCP,
			instanceTypes:  instancetypes.GCPInstanceTypes,
			expectedResult: true,
		},
		"put gcp when azure is set": {
			provider:       cloudprovider.Azure,
			instanceTypes:  instancetypes.GCPInstanceTypes,
			expectedResult: false,
		},
		"put gcp when azure is set with CVMs disabled": {
			provider:       cloudprovider.Azure,
			instanceTypes:  instancetypes.GCPInstanceTypes,
			nonCVMsAllowed: true,
			expectedResult: false,
		},
		"put azure when gcp is set": {
			provider:       cloudprovider.GCP,
			instanceTypes:  instancetypes.AzureCVMInstanceTypes,
			expectedResult: false,
		},
		"put azure when gcp is set with CVMs disabled": {
			provider:       cloudprovider.GCP,
			instanceTypes:  instancetypes.AzureTrustedLaunchInstanceTypes,
			nonCVMsAllowed: true,
			expectedResult: false,
		},
		// Testing every possible instance type for AWS is not feasible, so we just test a few based on known supported / unsupported families
		// Also serves as a test for checkIfInstanceInValidAWSFamilys
		"aws two valid instances": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"c5.xlarge", "c5a.2xlarge", "c5a.16xlarge", "u-12tb1.112xlarge"},
			expectedResult: true,
		},
		"aws one valid instance one with too little vCPUs": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"c5.medium"},
			expectedResult: false,
		},
		"aws graviton sub-family unsupported": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"m6g.xlarge", "r6g.2xlarge", "x2gd.xlarge", "g5g.8xlarge"},
			expectedResult: false,
		},
		"aws combined two valid instances as one string": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"c5.xlarge, c5a.2xlarge"},
			expectedResult: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			for _, instanceType := range tc.instanceTypes {
				assert.Equal(tc.expectedResult, validInstanceTypeForProvider(instanceType, tc.nonCVMsAllowed, tc.provider), instanceType)
			}
		})
	}
}

func TestIsDebugCluster(t *testing.T) {
	testCases := map[string]struct {
		config         *Config
		prepareConfig  func(*Config)
		expectedResult bool
	}{
		"empty config": {
			config:         &Config{},
			expectedResult: false,
		},
		"default config": {
			config:         DefaultWithPanic(),
			expectedResult: false,
		},
		"enabled": {
			config: DefaultWithPanic(),
			prepareConfig: func(conf *Config) {
				*conf.DebugCluster = true
			},
			expectedResult: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			if tc.prepareConfig != nil {
				tc.prepareConfig(tc.config)
			}
			assert.Equal(tc.expectedResult, tc.config.IsDebugCluster())
		})
	}
}

func TestValidateProvider(t *testing.T) {
	testCases := map[string]struct {
		provider         ProviderConfig
		wantErr          bool
		expectedErrorTag string
	}{
		"empty, should trigger no provider error": {
			provider:         ProviderConfig{},
			wantErr:          true,
			expectedErrorTag: "no_provider",
		},
		"azure only, should be okay": {
			provider: ProviderConfig{
				Azure: &AzureConfig{},
			},
			wantErr: false,
		},
		"gcp only, should be okay": {
			provider: ProviderConfig{
				GCP: &GCPConfig{},
			},
			wantErr: false,
		},
		"qemu only, should be okay": {
			provider: ProviderConfig{
				QEMU: &QEMUConfig{},
			},
			wantErr: false,
		},
		"azure and gcp, should trigger multiple provider error": {
			provider: ProviderConfig{
				Azure: &AzureConfig{},
				GCP:   &GCPConfig{},
			},
			wantErr:          true,
			expectedErrorTag: "more_than_one_provider",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			v := validator.New()
			trans := ut.New(en.New()).GetFallback()

			conf := DefaultWithPanic()
			conf.Provider = tc.provider

			v.RegisterStructValidation(validateProvider, ProviderConfig{})
			err := v.StructPartial(tc.provider)

			// Register provider validation error types.
			// Make sure the tags and expected strings below are in sync with the actual implementation.
			require.NoError(v.RegisterTranslation("no_provider", trans, registerNoProviderError, translateNoProviderError))
			require.NoError(v.RegisterTranslation("more_than_one_provider", trans, registerMoreThanOneProviderError, conf.translateMoreThanOneProviderError))

			// Continue if no error is expected.
			if !tc.wantErr {
				assert.NoError(err)
				return
			}

			// Validate if the error was identified correctly.
			require.NotNil(err)
			assert.Error(err)
			assert.Contains(err.Error(), tc.expectedErrorTag)

			// Check if error translation works correctly.
			validationErr := err.(validator.ValidationErrors)
			translatedErr := validationErr.Translate(trans)

			// The translator does not seem to export a list of available translations or for a specific field.
			// So we need to hardcode expected strings. Needs to be in sync with implementation.
			switch tc.expectedErrorTag {
			case "no_provider":
				assert.Contains(translatedErr["ProviderConfig.Provider"], "No provider has been defined")
			case "more_than_one_provider":
				assert.Contains(translatedErr["ProviderConfig.Provider"], "Only one provider can be defined")
			}
		})
	}
}

func TestConfigVersionCompatibility(t *testing.T) {
	t.Skip() // TODO(daniel-weisse): re-enable and re-write for config v3
	testCases := map[string]struct {
		config         string
		expectedConfig *Config
	}{
		"config v2 gcp": {
			config: "testdata/configGCPV2.yaml",
			expectedConfig: &Config{
				Version:           "v2",
				Image:             "v2.5.0",
				StateDiskSizeGB:   16,
				KubernetesVersion: "1.23",
				DebugCluster:      toPtr(false),
				Provider: ProviderConfig{
					GCP: &GCPConfig{
						Project:               "project-12345",
						Region:                "europe-west3",
						Zone:                  "europe-west3-b",
						ServiceAccountKeyPath: "serviceAccountKey.json",
						InstanceType:          "n2d-standard-4",
						StateDiskType:         "pd-ssd",
						DeployCSIDriver:       toPtr(true),
					},
				},
			},
		},
		"config v2 aws": {
			config: "testdata/configAWSV2.yaml",
			expectedConfig: &Config{
				Version:           "v2",
				Image:             "v2.5.0",
				StateDiskSizeGB:   16,
				KubernetesVersion: "1.23",
				DebugCluster:      toPtr(false),
				Provider: ProviderConfig{
					AWS: &AWSConfig{
						Region:                 "us-east-2",
						Zone:                   "us-east-2a",
						InstanceType:           "c5.xlarge",
						StateDiskType:          "gp2",
						IAMProfileControlPlane: "control_plane_instance_profile",
						IAMProfileWorkerNodes:  "node_instance_profile",
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fileHandler := file.NewHandler(afero.NewOsFs())

			config, err := fromFile(fileHandler, tc.config)

			assert.NoError(err)
			assert.Equal(tc.expectedConfig, config)
		})
	}
}
