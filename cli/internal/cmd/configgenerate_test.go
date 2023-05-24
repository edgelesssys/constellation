/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

func TestConfigGenerateKubernetesVersion(t *testing.T) {
	testCases := map[string]struct {
		version string
		wantErr bool
	}{
		"success": {
			version: semver.MajorMinor(string(versions.Default)),
		},
		"no semver": {
			version: "asdf",
			wantErr: true,
		},
		"not supported": {
			version: "1111",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cmd := newConfigGenerateCmd()
			err := cmd.Flags().Set("kubernetes", tc.version)
			require.NoError(err)

			cg := &configGenerateCmd{log: logger.NewTest(t)}
			err = cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, "")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestConfigGenerateDefault(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, ""))

	var readConfig config.Config
	err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
	assert.NoError(err)
	assert.Equal(*config.Default(), readConfig)
}

func TestConfigGenerateDefaultProviderSpecific(t *testing.T) {
	providers := []cloudprovider.Provider{
		cloudprovider.AWS,
		cloudprovider.Azure,
		cloudprovider.GCP,
		cloudprovider.OpenStack,
	}

	for _, provider := range providers {
		t.Run(provider.String(), func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cmd := newConfigGenerateCmd()

			wantConf := config.Default()
			wantConf.RemoveProviderAndAttestationExcept(provider)

			cg := &configGenerateCmd{log: logger.NewTest(t)}
			require.NoError(cg.configGenerate(cmd, fileHandler, provider, ""))

			var readConfig config.Config
			err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
			assert.NoError(err)
			assert.Equal(*wantConf, readConfig)
		})
	}
}

func TestConfigGenerateWithStackIt(t *testing.T) {
	openStackProviders := []string{"stackit"}

	for _, openStackProvider := range openStackProviders {
		t.Run(openStackProvider, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cmd := newConfigGenerateCmd()

			wantConf := config.Default().WithOpenStackProviderDefaults(openStackProvider)
			wantConf.RemoveProviderAndAttestationExcept(cloudprovider.OpenStack)

			cg := &configGenerateCmd{log: logger.NewTest(t)}
			require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.OpenStack, openStackProvider))

			var readConfig config.Config
			err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
			assert.NoError(err)
			assert.Equal(*wantConf, readConfig)
		})
	}
}

func TestConfigGenerateDefaultExists(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	require.NoError(fileHandler.Write(constants.ConfigFilename, []byte("foobar"), file.OptNone))
	cmd := newConfigGenerateCmd()

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.Error(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, ""))
}

func TestConfigGenerateFileFlagRemoved(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()
	cmd.ResetFlags()

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.Error(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, ""))
}

func TestConfigGenerateStdOut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())

	var outBuffer bytes.Buffer
	cmd := newConfigGenerateCmd()
	cmd.SetOut(&outBuffer)
	require.NoError(cmd.Flags().Set("file", "-"))

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, ""))

	var readConfig config.Config
	require.NoError(yaml.NewDecoder(&outBuffer).Decode(&readConfig))

	assert.Equal(*config.Default(), readConfig)
}

func TestNoValidProviderAttestationCombination(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		provider    cloudprovider.Provider
		attestation variant.Variant
	}{
		{cloudprovider.Azure, variant.AWSNitroTPM{}},
		{cloudprovider.AWS, variant.AzureTrustedLaunch{}},
		{cloudprovider.GCP, variant.AWSNitroTPM{}},
		{cloudprovider.QEMU, variant.GCPSEVES{}},
		{cloudprovider.OpenStack, variant.AWSNitroTPM{}},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			_, err := createConfigWithAttestationType(test.provider, "", test.attestation)
			assert.Error(err)
		})
	}
}

func TestValidProviderAttestationCombination(t *testing.T) {
	defaultAttestation := config.Default().Attestation
	tests := []struct {
		provider    cloudprovider.Provider
		attestation variant.Variant
		expected    config.AttestationConfig
	}{
		{
			cloudprovider.Azure,
			variant.AzureTrustedLaunch{},
			config.AttestationConfig{AzureTrustedLaunch: defaultAttestation.AzureTrustedLaunch},
		},
		{
			cloudprovider.Azure,
			variant.AzureSEVSNP{},
			config.AttestationConfig{AzureSEVSNP: defaultAttestation.AzureSEVSNP},
		},

		{
			cloudprovider.AWS,
			variant.AWSNitroTPM{},
			config.AttestationConfig{AWSNitroTPM: defaultAttestation.AWSNitroTPM},
		},
		{
			cloudprovider.GCP,
			variant.GCPSEVES{},
			config.AttestationConfig{GCPSEVES: defaultAttestation.GCPSEVES},
		},

		{
			cloudprovider.QEMU,
			variant.QEMUVTPM{},
			config.AttestationConfig{QEMUVTPM: defaultAttestation.QEMUVTPM},
		},
		{
			cloudprovider.OpenStack,
			variant.QEMUVTPM{},
			config.AttestationConfig{QEMUVTPM: defaultAttestation.QEMUVTPM},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Provider:%s,Attestation:%s", test.provider, test.attestation), func(t *testing.T) {
			sut, err := createConfigWithAttestationType(test.provider, "", test.attestation)
			assert := assert.New(t)
			assert.NoError(err)
			assert.Equal(test.expected, sut.Attestation)
		})
	}
}

func TestAttestationArgument(t *testing.T) {
	defaultAttestation := config.Default().Attestation
	tests := []struct {
		name        string
		provider    cloudprovider.Provider
		expectErr   bool
		expectedCfg config.AttestationConfig
		setFlag     func(*cobra.Command) error
	}{
		{
			name:      "InvalidAttestationArgument",
			provider:  cloudprovider.Unknown,
			expectErr: true,
			setFlag: func(cmd *cobra.Command) error {
				return cmd.Flags().Set("attestation", "unknown")
			},
		},
		{
			name:      "ValidAttestationArgument",
			provider:  cloudprovider.Azure,
			expectErr: false,
			setFlag: func(cmd *cobra.Command) error {
				return cmd.Flags().Set("attestation", "azure-trustedlaunch")
			},
			expectedCfg: config.AttestationConfig{AzureTrustedLaunch: defaultAttestation.AzureTrustedLaunch},
		},
		{
			name:      "WithoutAttestationArgument",
			provider:  cloudprovider.Azure,
			expectErr: false,
			setFlag: func(cmd *cobra.Command) error {
				return nil
			},
			expectedCfg: config.AttestationConfig{AzureSEVSNP: defaultAttestation.AzureSEVSNP},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require := assert.New(t)
			assert := assert.New(t)

			cmd := newConfigGenerateCmd()
			require.NoError(test.setFlag(cmd))

			fileHandler := file.NewHandler(afero.NewMemMapFs())

			cg := &configGenerateCmd{log: logger.NewTest(t)}
			err := cg.configGenerate(cmd, fileHandler, test.provider, "")
			if test.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				var readConfig config.Config
				require.NoError(fileHandler.ReadYAML(constants.ConfigFilename, &readConfig))

				assert.Equal(test.expectedCfg, readConfig.Attestation)
			}
		})
	}
}
