/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
)

func TestParseKubernetesVersion(t *testing.T) {
	testCases := map[string]struct {
		version string
		wantErr bool
	}{
		"default version": {
			version: "",
		},
		"without v prefix": {
			version: strings.TrimPrefix(string(versions.Default), "v"),
		},
		"K8s version without patch version": {
			version: semver.MajorMinor(string(versions.Default)),
		},
		"K8s version with patch version": {
			version: string(versions.Default),
		},
		"K8s version with invalid patch version": {
			version: func() string {
				s := string(versions.Default)
				return s[:len(s)-1] + "99"
			}(),
			wantErr: true,
		},
		"outdated K8s version": {
			version: "v1.0.0",
			wantErr: true,
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

			flags := newConfigGenerateCmd().Flags()
			if tc.version != "" {
				require.NoError(flags.Set("kubernetes", tc.version))
			}

			version, err := parseK8sFlag(flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(versions.Default, version)
		})
	}
}

func TestConfigGenerateDefault(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()

	cg := &configGenerateCmd{
		log: logger.NewTest(t),
		flags: generateFlags{
			attestationVariant: variant.Dummy{},
			k8sVersion:         versions.Default,
		},
	}
	require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, ""))

	var readConfig config.Config
	err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
	assert.NoError(err)
	assert.Equal(*config.Default(), readConfig)

	_, err = state.ReadFromFile(fileHandler, constants.StateFilename)
	assert.NoError(err)
}

func TestConfigGenerateDefaultProviderSpecific(t *testing.T) {
	testCases := map[string]struct {
		provider    cloudprovider.Provider
		rawProvider string
	}{
		"aws": {
			provider: cloudprovider.AWS,
		},
		"azure": {
			provider: cloudprovider.Azure,
		},
		"gcp": {
			provider: cloudprovider.GCP,
		},
		"openstack": {
			provider: cloudprovider.OpenStack,
		},
		"stackit": {
			provider:    cloudprovider.OpenStack,
			rawProvider: "stackit",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cmd := newConfigGenerateCmd()

			wantConf := config.Default().WithOpenStackProviderDefaults(tc.rawProvider)
			wantConf.RemoveProviderAndAttestationExcept(tc.provider)

			cg := &configGenerateCmd{
				log: logger.NewTest(t),
				flags: generateFlags{
					attestationVariant: variant.Dummy{},
					k8sVersion:         versions.Default,
				},
			}
			require.NoError(cg.configGenerate(cmd, fileHandler, tc.provider, tc.rawProvider))

			var readConfig config.Config
			err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
			assert.NoError(err)
			assert.Equal(*wantConf, readConfig)

			stateFile, err := state.ReadFromFile(fileHandler, constants.StateFilename)
			assert.NoError(err)
			switch tc.provider {
			case cloudprovider.GCP:
				assert.NotNil(stateFile.Infrastructure.GCP)
			case cloudprovider.Azure:
				assert.NotNil(stateFile.Infrastructure.Azure)
			}
		})
	}
}

func TestConfigGenerateDefaultExists(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	require.NoError(fileHandler.Write(constants.ConfigFilename, []byte("foobar"), file.OptNone))
	cmd := newConfigGenerateCmd()

	cg := &configGenerateCmd{
		log:   logger.NewTest(t),
		flags: generateFlags{attestationVariant: variant.Dummy{}},
	}
	require.Error(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown, ""))
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
			_, err := createConfigWithAttestationVariant(test.provider, "", test.attestation)
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
			variant.AWSSEVSNP{},
			config.AttestationConfig{AWSSEVSNP: defaultAttestation.AWSSEVSNP},
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
			sut, err := createConfigWithAttestationVariant(test.provider, "", test.attestation)
			assert := assert.New(t)
			assert.NoError(err)
			assert.Equal(test.expected, sut.Attestation)
		})
	}
}

func TestParseAttestationFlag(t *testing.T) {
	testCases := map[string]struct {
		wantErr         bool
		attestationFlag string
		wantVariant     variant.Variant
	}{
		"invalid": {
			wantErr:         true,
			attestationFlag: "unknown",
		},
		"AzureTrustedLaunch": {
			attestationFlag: "azure-trustedlaunch",
			wantVariant:     variant.AzureTrustedLaunch{},
		},
		"AzureSEVSNP": {
			attestationFlag: "azure-sev-snp",
			wantVariant:     variant.AzureSEVSNP{},
		},
		"AWSSEVSNP": {
			attestationFlag: "aws-sev-snp",
			wantVariant:     variant.AWSSEVSNP{},
		},
		"AWSNitroTPM": {
			attestationFlag: "aws-nitro-tpm",
			wantVariant:     variant.AWSNitroTPM{},
		},
		"GCPSEVES": {
			attestationFlag: "gcp-sev-es",
			wantVariant:     variant.GCPSEVES{},
		},
		"QEMUVTPM": {
			attestationFlag: "qemu-vtpm",
			wantVariant:     variant.QEMUVTPM{},
		},
		"no flag": {
			wantVariant: variant.Dummy{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			cmd := newConfigGenerateCmd()
			if tc.attestationFlag != "" {
				require.NoError(cmd.Flags().Set("attestation", tc.attestationFlag))
			}

			attestation, err := parseAttestationFlag(cmd.Flags())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.True(tc.wantVariant.Equal(attestation))
		})
	}
}
