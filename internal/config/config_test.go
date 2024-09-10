/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"context"
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

	"github.com/edgelesssys/constellation/v2/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	gosemver "golang.org/x/mod/semver"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	def := Default()
	assert.NotNil(def)
}

func TestDefaultConfigMarshalsLatestVersion(t *testing.T) {
	conf := Default()
	bt, err := yaml.Marshal(conf)
	require := require.New(t)
	require.NoError(err)

	var mp configMap
	require.NoError(yaml.Unmarshal(bt, &mp))
	assert := assert.New(t)
	assert.Equal("latest", mp.getSEVSNPVersion("microcodeVersion"))
	assert.Equal("latest", mp.getSEVSNPVersion("teeVersion"))
	assert.Equal("latest", mp.getSEVSNPVersion("snpVersion"))
	assert.Equal("latest", mp.getSEVSNPVersion("bootloaderVersion"))
}

func TestGetAttestationConfigMarshalsNumericalVersion(t *testing.T) {
	conf := Default()
	conf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)

	attestationCfg := conf.GetAttestationConfig()
	bt, err := yaml.Marshal(attestationCfg)
	require := require.New(t)
	require.NoError(err)

	var mp map[string]interface{}
	require.NoError(yaml.Unmarshal(bt, &mp))
	assert := assert.New(t)
	assert.EqualValues(placeholderVersionValue[uint8](), mp["microcodeVersion"])
	assert.EqualValues(placeholderVersionValue[uint8](), mp["teeVersion"])
	assert.EqualValues(placeholderVersionValue[uint8](), mp["snpVersion"])
	assert.EqualValues(placeholderVersionValue[uint8](), mp["bootloaderVersion"])
}

func TestNew(t *testing.T) {
	testCases := map[string]struct {
		config        configMap
		configName    string
		wantResult    *Config
		wantErr       bool
		wantedErrType error
	}{
		"Azure SEV-SNP: mix of Latest and uint as version value in file correctly sets latest versions values": {
			config: func() configMap {
				conf := Default() // default configures latest version
				modifyConfigForAzureToPassValidate(conf)
				m := getConfigAsMap(conf, t)
				m.setSEVSNPVersion("microcodeVersion", "Latest") // check uppercase also works
				m.setSEVSNPVersion("teeVersion", 2)
				m.setSEVSNPVersion("bootloaderVersion", 1)
				return m
			}(),

			configName: constants.ConfigFilename,
			wantResult: func() *Config {
				conf := Default()
				modifyConfigForAzureToPassValidate(conf)
				conf.Attestation.AzureSEVSNP.MicrocodeVersion = AttestationVersion[uint8]{
					Value:      testCfg.Microcode,
					WantLatest: true,
				}
				conf.Attestation.AzureSEVSNP.TEEVersion = AttestationVersion[uint8]{
					Value:      2,
					WantLatest: false,
				}
				conf.Attestation.AzureSEVSNP.BootloaderVersion = AttestationVersion[uint8]{
					Value:      1,
					WantLatest: false,
				}
				conf.Attestation.AzureSEVSNP.SNPVersion = AttestationVersion[uint8]{
					Value:      testCfg.SNP,
					WantLatest: true,
				}
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
			result, err := New(fileHandler, tc.configName, stubAttestationFetcher{}, false)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantResult, result)
		})
	}
}

func modifyConfigForAzureToPassValidate(c *Config) {
	c.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
	c.Image = constants.BinaryVersion().String()
	c.Provider.Azure.SubscriptionID = "11111111-1111-1111-1111-111111111111"
	c.Provider.Azure.TenantID = "11111111-1111-1111-1111-111111111111"
	c.Provider.Azure.Location = "westus"
	c.Provider.Azure.ResourceGroup = "test"
	c.Provider.Azure.UserAssignedIdentity = "/subscriptions/11111111-1111-1111-1111-111111111111/resourcegroups/constellation-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-identity"
	c.Attestation.AzureSEVSNP.Measurements = measurements.M{
		0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
	}
	c.NodeGroups = map[string]NodeGroup{
		constants.ControlPlaneDefault: {
			Role:            "control-plane",
			Zone:            "",
			InstanceType:    "Standard_DC4as_v5",
			StateDiskSizeGB: 30,
			StateDiskType:   "StandardSSD_LRS",
			InitialCount:    3,
		},
		constants.WorkerDefault: {
			Role:            "worker",
			Zone:            "",
			InstanceType:    "Standard_DC4as_v5",
			StateDiskSizeGB: 30,
			StateDiskType:   "StandardSSD_LRS",
			InitialCount:    3,
		},
	}
}

func TestReadConfigFile(t *testing.T) {
	testCases := map[string]struct {
		config        configMap
		configName    string
		wantResult    *Config
		wantErr       bool
		wantedErrType error
	}{
		"refuse invalid version value": {
			config: func() configMap {
				conf := Default()
				m := getConfigAsMap(conf, t)
				m.setSEVSNPVersion("microcodeVersion", "1a")
				return m
			}(),
			configName: constants.ConfigFilename,
			wantErr:    true,
		},
		"outdated k8s patch version is allowed": {
			config: func() configMap {
				conf := Default()
				ver, err := semver.New(versions.SupportedK8sVersions()[0])
				require.NoError(t, err)
				conf.KubernetesVersion = versions.ValidK8sVersion(semver.NewFromInt(ver.Major(), ver.Minor(), ver.Patch()-1, "").String())
				m := getConfigAsMap(conf, t)
				return m
			}(),
			wantResult: func() *Config {
				conf := Default()
				ver, err := semver.New(versions.SupportedK8sVersions()[0])
				require.NoError(t, err)
				conf.KubernetesVersion = versions.ValidK8sVersion(semver.NewFromInt(ver.Major(), ver.Minor(), ver.Patch()-1, "").String())
				return conf
			}(),
			configName: constants.ConfigFilename,
		},
		"outdated k8s version is not allowed": {
			config: func() configMap {
				conf := Default()
				conf.KubernetesVersion = versions.ValidK8sVersion("v1.0.0")
				m := getConfigAsMap(conf, t)
				return m
			}(),
			wantErr:    true,
			configName: constants.ConfigFilename,
		},
		"a k8s version without specified patch is not allowed": {
			config: func() configMap {
				conf := Default()
				conf.KubernetesVersion = versions.ValidK8sVersion(gosemver.MajorMinor(string(versions.Default)))
				m := getConfigAsMap(conf, t)
				return m
			}(),
			wantErr:    true,
			configName: constants.ConfigFilename,
		},
		"error on entering app client id": {
			config: func() configMap {
				conf := Default()
				m := getConfigAsMap(conf, t)
				m.setAzureProvider("appClientID", "3ea4bdc1-1cc1-4237-ae78-0831eff3491e")
				return m
			}(),
			configName:    constants.ConfigFilename,
			wantedErrType: &UnsupportedAppRegistrationError{},
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
			if tc.wantedErrType != nil {
				assert.ErrorIs(err, tc.wantedErrType)
				return
			}
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantResult, result)
		})
	}
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

func TestValidate(t *testing.T) {
	const defaultErrCount = 33 // expect this number of error messages by default because user-specific values are not set and multiple providers are defined by default
	const azErrCount = 7
	const awsErrCount = 8
	const gcpErrCount = 8

	// TODO(AB#3132,3u13r): refactor config validation tests
	// Note that the `cnf.Image = ""` is a hack to align `bazel test` with `go test` behavior
	// since first does version stamping.
	testCases := map[string]struct {
		cnf          *Config
		wantErr      bool
		wantErrCount int
	}{
		"default config is not valid": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Image = ""
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount,
		},
		"microservices violate version drift": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Image = ""
				cliVersion := constants.BinaryVersion()
				cnf.MicroserviceVersion = semver.NewFromInt(cliVersion.Major()+2, cliVersion.Minor(), cliVersion.Patch(), "")
				return cnf
			}(),
			wantErr: true,
			// This is a very different value from the other error counts because of the way we are checking MicroserviceVersions.
			wantErrCount: 1,
		},
		"v0 is one error": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Image = ""
				cnf.Version = "v0"
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount + 1,
		},
		"v0 and long name are two errors": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Image = ""
				cnf.Version = "v0"
				cnf.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: defaultErrCount + 2,
		},
		"default Azure config is not valid": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Image = ""
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: azErrCount,
		},
		"Azure config with all required fields is valid": {
			cnf: func() *Config {
				cnf := Default()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
				cnf.Image = constants.BinaryVersion().String()
				modifyConfigForAzureToPassValidate(cnf)
				return cnf
			}(),
		},
		"default AWS config is not valid": {
			cnf: func() *Config {
				cnf := Default()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: awsErrCount,
		},
		"AWS config with upper case name": {
			cnf: func() *Config {
				cnf := Default()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
				cnf.Name = "testAWS"
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: awsErrCount + 1,
		},
		"AWS config with correct region and zone format": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Provider.AWS.Region = "us-east-2"
				cnf.Provider.AWS.Zone = "us-east-2a"
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: awsErrCount - 4,
		},
		"AWS config with wrong region and zone format": {
			cnf: func() *Config {
				cnf := Default()
				cnf.Provider.AWS.Region = "us-west2"
				cnf.Provider.AWS.Zone = "a"
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: awsErrCount,
		},
		"default GCP config is not valid": {
			cnf: func() *Config {
				cnf := Default()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.GCP)
				cnf.Image = ""
				return cnf
			}(),
			wantErr:      true,
			wantErrCount: gcpErrCount,
		},

		"GCP config with all required fields is valid": {
			cnf: func() *Config {
				cnf := Default()
				cnf.RemoveProviderAndAttestationExcept(cloudprovider.GCP)
				cnf.Image = constants.BinaryVersion().String()
				gcp := cnf.Provider.GCP
				gcp.Region = "test-region"
				gcp.Project = "test-project"
				gcp.Zone = "test-zone"
				gcp.ServiceAccountKeyPath = "test-key-path"
				cnf.Provider = ProviderConfig{}
				cnf.Provider.GCP = gcp
				cnf.Attestation.GCPSEVSNP.Measurements = measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
				}
				cnf.NodeGroups = map[string]NodeGroup{
					constants.ControlPlaneDefault: {
						Role:            "control-plane",
						Zone:            "europe-west1-b",
						InstanceType:    "n2d-standard-4",
						StateDiskSizeGB: 30,
						StateDiskType:   "pd-ssd",
						InitialCount:    3,
					},
					constants.WorkerDefault: {
						Role:            "worker",
						Zone:            "europe-west1-b",
						InstanceType:    "n2d-standard-4",
						StateDiskSizeGB: 30,
						StateDiskType:   "pd-ssd",
						InitialCount:    3,
					},
				}
				return cnf
			}(),
		},
		"miniup default config is not valid because image and measurements are missing in OSS": {
			cnf: func() *Config {
				config, _ := MiniDefault()
				require.NotNil(t, config)
				return config
			}(),
			wantErr:      true,
			wantErrCount: 2,
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
				assert.Equalf(tc.wantErrCount, valErr.messagesCount(), "Got unexpected error count: %d: %s", valErr.messagesCount(), valErr.LongMessage())
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
	assert.False(Default().HasProvider(cloudprovider.Unknown))
	assert.True(Default().HasProvider(cloudprovider.Azure))
	assert.True(Default().HasProvider(cloudprovider.GCP))
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
		conf := Default()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.AWS)
		for k := range conf.Attestation.AWSSEVSNP.Measurements {
			delete(conf.Attestation.AWSSEVSNP.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.AWSSEVSNP.Measurements)
	}
	{ // Azure
		conf := Default()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
		for k := range conf.Attestation.AzureSEVSNP.Measurements {
			delete(conf.Attestation.AzureSEVSNP.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.AzureSEVSNP.Measurements)
	}
	{ // GCP
		conf := Default()
		conf.RemoveProviderAndAttestationExcept(cloudprovider.GCP)
		for k := range conf.Attestation.GCPSEVSNP.Measurements {
			delete(conf.Attestation.GCPSEVSNP.Measurements, k)
		}
		conf.UpdateMeasurements(newMeasurements)
		assert.Equal(newMeasurements, conf.Attestation.GCPSEVSNP.Measurements)
	}
	{ // QEMU
		conf := Default()
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
				conf := Default()
				conf.Image = "v0.0.0"
				return conf
			}(),
			want: true,
		},
		"branch image": {
			conf: func() *Config {
				conf := Default()
				conf.Image = "feat-x-vX.Y.Z-pre.0.yyyymmddhhmmss-abcdefabcdef"
				return conf
			}(),
			want: false,
		},
		"debug image": {
			conf: func() *Config {
				conf := Default()
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
		variant        variant.Variant
		instanceTypes  []string
		expectedResult bool
	}{
		"empty all": {
			variant:        variant.Dummy{},
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty aws": {
			variant:        variant.AWSSEVSNP{},
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty azure only CVMs": {
			variant:        variant.AzureSEVSNP{},
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty azure with non-CVMs": {
			variant:        variant.AzureTrustedLaunch{},
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"empty gcp": {
			variant:        variant.GCPSEVES{},
			instanceTypes:  []string{},
			expectedResult: false,
		},
		"azure only CVMs (SNP)": {
			variant:        variant.AzureSEVSNP{},
			instanceTypes:  instancetypes.AzureSNPInstanceTypes,
			expectedResult: true,
		},
		"azure only CVMs (TDX)": {
			variant:        variant.AzureTDX{},
			instanceTypes:  instancetypes.AzureTDXInstanceTypes,
			expectedResult: true,
		},
		"azure trusted launch VMs": {
			variant:        variant.AzureTrustedLaunch{},
			instanceTypes:  instancetypes.AzureTrustedLaunchInstanceTypes,
			expectedResult: true,
		},
		"gcp": {
			variant:        variant.GCPSEVES{},
			instanceTypes:  instancetypes.GCPInstanceTypes,
			expectedResult: true,
		},
		"gcp sev-snp": {
			variant:        variant.GCPSEVSNP{},
			instanceTypes:  instancetypes.GCPInstanceTypes,
			expectedResult: true,
		},
		"put gcp when azure is set": {
			variant:        variant.AzureSEVSNP{},
			instanceTypes:  instancetypes.GCPInstanceTypes,
			expectedResult: false,
		},
		"put azure when gcp is set": {
			variant:        variant.GCPSEVES{},
			instanceTypes:  instancetypes.AzureSNPInstanceTypes,
			expectedResult: false,
		},
		// Testing every possible instance type for AWS is not feasible, so we just test a few based on known supported / unsupported families
		// Also serves as a test for checkIfInstanceInValidAWSFamilys
		"aws two valid instances": {
			variant:        variant.AWSSEVSNP{},
			instanceTypes:  []string{"c5.xlarge", "c5a.2xlarge", "c5a.16xlarge", "u-12tb1.112xlarge"},
			expectedResult: false, // False because 2 two of the instances are not valid
		},
		"aws one valid instance one with too little vCPUs": {
			variant:        variant.AWSSEVSNP{},
			instanceTypes:  []string{"c5.medium"},
			expectedResult: false,
		},
		"aws graviton sub-family unsupported": {
			variant:        variant.AWSSEVSNP{},
			instanceTypes:  []string{"m6g.xlarge", "r6g.2xlarge", "x2gd.xlarge", "g5g.8xlarge"},
			expectedResult: false,
		},
		"aws combined two valid instances as one string": {
			variant:        variant.AWSSEVSNP{},
			instanceTypes:  []string{"c5.xlarge, c5a.2xlarge"},
			expectedResult: false,
		},
		"aws only CVMs": {
			variant:        variant.AWSSEVSNP{},
			instanceTypes:  []string{"c6a.xlarge", "m6a.xlarge", "r6a.xlarge"},
			expectedResult: true,
		},
		"aws nitroTPM VMs": {
			variant:        variant.AWSNitroTPM{},
			instanceTypes:  []string{"c5.xlarge", "c5a.2xlarge", "c5a.16xlarge", "u-12tb1.112xlarge"},
			expectedResult: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			for _, instanceType := range tc.instanceTypes {
				assert.Equal(
					tc.expectedResult, validInstanceTypeForProvider(instanceType, tc.variant),
					instanceType,
				)
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

func TestConfigVersionCompatibility(t *testing.T) {
	t.Skip() // TODO(daniel-weisse): re-enable and re-write for config v4
	testCases := map[string]struct {
		config         string
		expectedConfig *Config
	}{
		"config v2 gcp": {
			config: "testdata/configGCPV2.yaml",
			expectedConfig: &Config{
				Version:           "v2",
				Image:             "v2.5.0",
				KubernetesVersion: "1.23",
				DebugCluster:      toPtr(false),
				Provider: ProviderConfig{
					GCP: &GCPConfig{
						Project:               "project-12345",
						Region:                "europe-west3",
						Zone:                  "europe-west3-b",
						ServiceAccountKeyPath: "serviceAccountKey.json",
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
				KubernetesVersion: "1.23",
				DebugCluster:      toPtr(false),
				Provider: ProviderConfig{
					AWS: &AWSConfig{
						Region:                 "us-east-2",
						Zone:                   "us-east-2a",
						IAMProfileControlPlane: "control_plane_instance_profile_name",
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

func TestIsDebugImage(t *testing.T) {
	cases := map[string]struct {
		image    string
		expected bool
	}{
		"debug image":   {"ref/test/stream/debug/v2.9.0-pre.0.20230613084544-eeea7b1f56f4", true},
		"release image": {"v2.8.0", false},
		"empty image":   {"", false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &Config{Image: tc.image}
			assert.Equal(t, tc.expected, c.IsNamedLikeDebugImage())
		})
	}
}

func TestIsAppClientIDError(t *testing.T) {
	testCases := map[string]struct {
		err      error
		expected bool
	}{
		"yaml.Error with appClientID error": {
			err: &yaml.TypeError{
				Errors: []string{
					"invalid value for appClientID",
					"another error",
				},
			},
			expected: true,
		},
		"yaml.Error without appClientID error": {
			err: &yaml.TypeError{
				Errors: []string{
					"invalid value for something else",
					"another error",
				},
			},
			expected: false,
		},
		"other error": {
			err:      errors.New("appClientID but other error type"),
			expected: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.expected, isAppClientIDError(tc.err))
		})
	}
}

// configMap is used to un-/marshal the config as an unstructured map.
type configMap map[string]interface{}

func (c configMap) setSEVSNPVersion(versionType string, value interface{}) {
	c["attestation"].(configMap)["azureSEVSNP"].(configMap)[versionType] = value
}

func (c configMap) setAzureProvider(azureProviderField string, value interface{}) {
	c["provider"].(configMap)["azure"].(configMap)[azureProviderField] = value
}

func (c configMap) getSEVSNPVersion(versionType string) interface{} {
	return c["attestation"].(configMap)["azureSEVSNP"].(configMap)[versionType]
}

// getConfigAsMap returns a map of the config.
func getConfigAsMap(conf *Config, t *testing.T) (res configMap) {
	bytes, err := yaml.Marshal(&conf)
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(bytes, &res); err != nil {
		t.Fatal(err)
	}
	return
}

type stubAttestationFetcher struct{}

func (f stubAttestationFetcher) FetchLatestVersion(_ context.Context, _ attestationconfig.Variant) (attestationconfig.Entry, error) {
	return attestationconfig.Entry{
		SEVSNPVersion: testCfg,
	}, nil
}

var testCfg = attestationconfig.SEVSNPVersion{
	Microcode:  93,
	TEE:        0,
	SNP:        6,
	Bootloader: 2,
}
