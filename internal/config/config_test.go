package config

import (
	"reflect"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
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
			wantResult: Default(),
		},
		"err when path not exist": {
			config:     nil,
			configName: "wrong-name.yaml",
			wantErr:    true,
		},
		"custom config from default file": {
			config: &Config{
				Version:                 Version1,
				AutoscalingNodeGroupMin: 42,
				AutoscalingNodeGroupMax: 1337,
			},
			configName: constants.ConfigFilename,
			wantResult: &Config{
				Version:                 Version1,
				AutoscalingNodeGroupMin: 42,
				AutoscalingNodeGroupMax: 1337,
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

func TestFromFileStrictErrors(t *testing.T) {
	testCases := map[string]struct {
		yamlConfig string
		wantErr    bool
	}{
		"valid config": {
			yamlConfig: `
			autoscalingNodeGroupMin: 5
			autoscalingNodeGroupMax: 10
			stateDisksizeGB: 25
			`,
		},
		"typo": {
			yamlConfig: `
			autoscalingNodeGroupMini: 5
			autoscalingNodeGroupMax: 10
			stateDisksizeGB: 25
			`,
			wantErr: true,
		},
		"unsupported version": {
			yamlConfig: `
			version: v5
			autoscalingNodeGroupMin: 1
			autoscalingNodeGroupMax: 10
			stateDisksizeGB: 30
			`,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			err := fileHandler.Write(constants.ConfigFilename, []byte(tc.yamlConfig), file.OptNone)
			assert.NoError(err)

			_, err = FromFile(fileHandler, constants.ConfigFilename)
			assert.Error(err)
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := map[string]struct {
		cnf          *Config
		wantMsgCount int
	}{
		"default config is valid": {
			cnf:          Default(),
			wantMsgCount: 0,
		},
		"config with 1 error": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Version = "v0"
				return cnf
			}(),
			wantMsgCount: 1,
		},
		"config with 2 errors": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Version = "v0"
				cnf.StateDiskSizeGB = -1
				return cnf
			}(),
			wantMsgCount: 2,
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
		wantAzure    *AzureConfig
		wantGCP      *GCPConfig
		wantQEMU     *QEMUConfig
	}{
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
	assert.Len(UserKeyDoc.Fields, reflect.ValueOf(UserKey{}).NumField(), updateMsg)
	assert.Len(FirewallRuleDoc.Fields, reflect.ValueOf(FirewallRule{}).NumField(), updateMsg)
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
