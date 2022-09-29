/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"reflect"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	def := Default()
	assert.NotNil(def)
}

func TestFromFile(t *testing.T) {
	testCases := map[string]struct {
		config     *Config
		configName string
		wantResult *Config
		wantErr    bool
	}{
		"default config from default file": {
			config:     Default(),
			configName: constants.ConfigFilename,
			wantResult: Default(),
		},
		"default config from different path": {
			config:     Default(),
			configName: "other-config.yaml",
			wantResult: Default(),
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
				Version: Version1,
			},
			configName: constants.ConfigFilename,
			wantResult: &Config{
				Version: Version1,
			},
		},
		"modify default config": {
			config: func() *Config {
				conf := Default()
				conf.Provider.GCP.Region = "eu-north1"
				conf.Provider.GCP.Zone = "eu-north1-a"
				return conf
			}(),
			configName: constants.ConfigFilename,
			wantResult: func() *Config {
				conf := Default()
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

			result, err := FromFile(fileHandler, tc.configName)

			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(tc.wantResult, result)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	const defaultMsgCount = 20 // expect this number of error messages by default because user-specific values are not set and multiple providers are defined by default

	testCases := map[string]struct {
		cnf          *Config
		wantMsgCount int
	}{
		"default config is valid": {
			cnf:          Default(),
			wantMsgCount: defaultMsgCount,
		},
		"config with 1 error": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Version = "v0"
				return cnf
			}(),
			wantMsgCount: defaultMsgCount + 1,
		},
		"config with 2 errors": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Version = "v0"
				cnf.StateDiskSizeGB = -1
				return cnf
			}(),
			wantMsgCount: defaultMsgCount + 2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			msgs, err := tc.cnf.Validate()
			require.NoError(err)
			assert.Len(msgs, tc.wantMsgCount)
		})
	}
}

func TestHasProvider(t *testing.T) {
	assert := assert.New(t)
	assert.False((&Config{}).HasProvider(cloudprovider.Unknown))
	assert.False((&Config{}).HasProvider(cloudprovider.Azure))
	assert.False((&Config{}).HasProvider(cloudprovider.GCP))
	assert.False((&Config{}).HasProvider(cloudprovider.QEMU))
	assert.False(Default().HasProvider(cloudprovider.Unknown))
	assert.True(Default().HasProvider(cloudprovider.Azure))
	assert.True(Default().HasProvider(cloudprovider.GCP))
	cnfWithAzure := Config{Provider: ProviderConfig{Azure: &AzureConfig{}}}
	assert.False(cnfWithAzure.HasProvider(cloudprovider.Unknown))
	assert.True(cnfWithAzure.HasProvider(cloudprovider.Azure))
	assert.False(cnfWithAzure.HasProvider(cloudprovider.GCP))
}

func TestImage(t *testing.T) {
	testCases := map[string]struct {
		cfg       *Config
		wantImage string
	}{
		"default aws": {
			cfg:       func() *Config { c := Default(); c.RemoveProviderExcept(cloudprovider.AWS); return c }(),
			wantImage: Default().Provider.AWS.Image,
		},
		"default azure": {
			cfg:       func() *Config { c := Default(); c.RemoveProviderExcept(cloudprovider.Azure); return c }(),
			wantImage: Default().Provider.Azure.Image,
		},
		"default gcp": {
			cfg:       func() *Config { c := Default(); c.RemoveProviderExcept(cloudprovider.GCP); return c }(),
			wantImage: Default().Provider.GCP.Image,
		},
		"default qemu": {
			cfg:       func() *Config { c := Default(); c.RemoveProviderExcept(cloudprovider.QEMU); return c }(),
			wantImage: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			image := tc.cfg.Image()
			assert.Equal(tc.wantImage, image)
		})
	}
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
			wantAWS:      Default().Provider.AWS,
		},
		"except azure": {
			removeExcept: cloudprovider.Azure,
			wantAzure:    Default().Provider.Azure,
		},
		"except gcp": {
			removeExcept: cloudprovider.GCP,
			wantGCP:      Default().Provider.GCP,
		},
		"except qemu": {
			removeExcept: cloudprovider.QEMU,
			wantQEMU:     Default().Provider.QEMU,
		},
		"unknown provider": {
			removeExcept: cloudprovider.Unknown,
			wantAWS:      Default().Provider.AWS,
			wantAzure:    Default().Provider.Azure,
			wantGCP:      Default().Provider.GCP,
			wantQEMU:     Default().Provider.QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			conf := Default()
			conf.RemoveProviderExcept(tc.removeExcept)

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
	assert.Len(UpgradeConfigDoc.Fields, reflect.ValueOf(UpgradeConfig{}).NumField(), updateMsg)
	assert.Len(UserKeyDoc.Fields, reflect.ValueOf(UserKey{}).NumField(), updateMsg)
	assert.Len(ProviderConfigDoc.Fields, reflect.ValueOf(ProviderConfig{}).NumField(), updateMsg)
	assert.Len(AzureConfigDoc.Fields, reflect.ValueOf(AzureConfig{}).NumField(), updateMsg)
	assert.Len(GCPConfigDoc.Fields, reflect.ValueOf(GCPConfig{}).NumField(), updateMsg)
	assert.Len(QEMUConfigDoc.Fields, reflect.ValueOf(QEMUConfig{}).NumField(), updateMsg)
}

func TestConfig_UpdateMeasurements(t *testing.T) {
	assert := assert.New(t)
	newMeasurements := Measurements{
		1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	}

	{ // AWS
		conf := Default()
		conf.RemoveProviderExcept(cloudprovider.AWS)
		for k := range conf.Provider.AWS.Measurements {
			delete(conf.Provider.AWS.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Provider.AWS.Measurements)
	}
	{ // Azure
		conf := Default()
		conf.RemoveProviderExcept(cloudprovider.Azure)
		for k := range conf.Provider.Azure.Measurements {
			delete(conf.Provider.Azure.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Provider.Azure.Measurements)
	}
	{ // GCP
		conf := Default()
		conf.RemoveProviderExcept(cloudprovider.GCP)
		for k := range conf.Provider.GCP.Measurements {
			delete(conf.Provider.GCP.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Provider.GCP.Measurements)
	}
	{ // QEMU
		conf := Default()
		conf.RemoveProviderExcept(cloudprovider.QEMU)
		for k := range conf.Provider.QEMU.Measurements {
			delete(conf.Provider.QEMU.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Provider.QEMU.Measurements)
	}
}

func TestConfig_IsImageDebug(t *testing.T) {
	testCases := map[string]struct {
		conf *Config
		want bool
	}{
		// TODO: Add AWS when we know the format of published images & debug images
		"gcp release": {
			conf: func() *Config {
				conf := Default()
				conf.RemoveProviderExcept(cloudprovider.GCP)
				conf.Provider.GCP.Image = "projects/constellation-images/global/images/constellation-v1-3-0"
				return conf
			}(),
			want: false,
		},
		"gcp debug": {
			conf: func() *Config {
				conf := Default()
				conf.RemoveProviderExcept(cloudprovider.GCP)
				conf.Provider.GCP.Image = "projects/constellation-images/global/images/constellation-20220812102023"
				return conf
			}(),
			want: true,
		},
		"azure release": {
			conf: func() *Config {
				conf := Default()
				conf.RemoveProviderExcept(cloudprovider.Azure)
				conf.Provider.Azure.Image = "/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/0.0.1"
				return conf
			}(),
			want: false,
		},
		"azure debug": {
			conf: func() *Config {
				conf := Default()
				conf.RemoveProviderExcept(cloudprovider.Azure)
				conf.Provider.Azure.Image = "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation_Debug/images/v1.4.0/versions/2022.0805.151600"
				return conf
			}(),
			want: true,
		},
		"empty config": {
			conf: &Config{},
			want: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.want, tc.conf.IsDebugImage())
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
		// Testing every possible instance type for AWS is not feasible, so we just test a few.
		// Also serves as a test for checkIfInstanceInValidAWSFamilys
		"aws two valid instances": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"t3.xlarge", "g5.2xlarge"},
			expectedResult: true,
		},
		"aws one valid metal instance": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"m6id.metal"},
			expectedResult: true,
		},
		"aws one valid instance one with too little vCPUs": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"t3.xlarge, t3.medium"},
			expectedResult: false,
		},
		"aws one valid instance one graviton": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"t3.xlarge, a1.xlarge"},
			expectedResult: false,
		},
		"aws combined two instances as one": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"t3.xlarge, t3.2xlarge"},
			expectedResult: false,
		},
		"aws explicit definition of a specific size (supported)": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"u-12tb1.112xlarge", "u-6tb1.metal"},
			expectedResult: true,
		},
		"aws explicit definition of a specific size (unsupported)": {
			provider:       cloudprovider.AWS,
			instanceTypes:  []string{"u-6tb1.92xlarge"}, // This instance type is made up just to test whether the check works explicitly on the full name, not just the family
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
			config:         Default(),
			expectedResult: false,
		},
		"enabled": {
			config: Default(),
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

			conf := Default()
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
