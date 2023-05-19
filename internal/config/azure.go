/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestationapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

// AzureSEVSNP is the configuration for Azure SEV-SNP attestation.
type AzureSEVSNP struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// description: |
	//   Lowest acceptable bootloader version.
	BootloaderVersion uint8 `json:"bootloaderVersion" yaml:"bootloaderVersion"`
	// description: |
	//   Lowest acceptable TEE version.
	TEEVersion uint8 `json:"teeVersion" yaml:"teeVersion"`
	// description: |
	//   Lowest acceptable SEV-SNP version.
	SNPVersion uint8 `json:"snpVersion" yaml:"snpVersion"`
	// description: |
	//   Lowest acceptable microcode version.
	MicrocodeVersion uint8 `json:"microcodeVersion" yaml:"microcodeVersion"`
	// description: |
	//   Configuration for validating the firmware signature.
	FirmwareSignerConfig SNPFirmwareSignerConfig `json:"firmwareSignerConfig" yaml:"firmwareSignerConfig"`
	// description: |
	//   AMD Root Key certificate used to verify the SEV-SNP certificate chain.
	AMDRootKey Certificate `json:"amdRootKey" yaml:"amdRootKey"`
}

// DefaultForAzureSEVSNP returns the default configuration for Azure SEV-SNP attestation.
// Version numbers are hard coded and should be updated with each new release.
func DefaultForAzureSEVSNP() (*AzureSEVSNP, error) {
	ctx := context.Background()
	version, err := attestationapi.GetAzureSEVSNPVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &AzureSEVSNP{
		Measurements:      measurements.DefaultsFor(cloudprovider.Azure, variant.AzureSEVSNP{}),
		BootloaderVersion: version.Bootloader,
		TEEVersion:        version.TEE,
		SNPVersion:        version.SNP,
		MicrocodeVersion:  version.Microcode,
		FirmwareSignerConfig: SNPFirmwareSignerConfig{
			AcceptedKeyDigests: idkeydigest.DefaultList(),
			EnforcementPolicy:  idkeydigest.MAAFallback,
		},
		// AMD root key. Received from the AMD Key Distribution System API (KDS).
		AMDRootKey: mustParsePEM(`-----BEGIN CERTIFICATE-----\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\nAFZEAwoKCQ==\n-----END CERTIFICATE-----\n`),
	}, nil
}

// GetVariant returns azure-sev-snp as the variant.
func (AzureSEVSNP) GetVariant() variant.Variant {
	return variant.AzureSEVSNP{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AzureSEVSNP) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AzureSEVSNP) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AzureSEVSNP) EqualTo(old AttestationCfg) (bool, error) {
	otherCfg, ok := old.(*AzureSEVSNP)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, old)
	}

	firmwareSignerCfgEqual := c.FirmwareSignerConfig.EqualTo(otherCfg.FirmwareSignerConfig)
	measurementsEqual := c.Measurements.EqualTo(otherCfg.Measurements)
	bootloaderEqual := c.BootloaderVersion == otherCfg.BootloaderVersion
	teeEqual := c.TEEVersion == otherCfg.TEEVersion
	snpEqual := c.SNPVersion == otherCfg.SNPVersion
	microcodeEqual := c.MicrocodeVersion == otherCfg.MicrocodeVersion
	rootKeyEqual := bytes.Equal(c.AMDRootKey.Raw, otherCfg.AMDRootKey.Raw)

	return firmwareSignerCfgEqual && measurementsEqual && bootloaderEqual && teeEqual && snpEqual && microcodeEqual && rootKeyEqual, nil
}

// UnmarshalYAML implements a custom unmarshaler to support setting "latest" as version.
func (a *AzureSEVSNP) UnmarshalYAML(unmarshal func(interface{}) error) error {
	aux := &fusedAzureSEVSNP{
		auxAzureSEVSNP: (*auxAzureSEVSNP)(a),
	}
	if err := unmarshal(aux); err != nil {
		return fmt.Errorf("failed to unmarshal AzureSEVSNP: %w", err)
	}
	a = (*AzureSEVSNP)(aux.auxAzureSEVSNP)

	versions, err := attestationapi.GetAzureSEVSNPVersion(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get AzureSEVSNP versions: %w", err)
	}
	for _, versionType := range []attestationapi.Type{attestationapi.Bootloader, attestationapi.TEE, attestationapi.SNP, attestationapi.Microcode} {
		if !convertLatestToNumber(a, versions, versionType, aux) {
			if err := convertStringToUint(a, versionType, aux); err != nil {
				return fmt.Errorf("failed to convert %s version to number: %w", versionType, err)
			}
		}
	}
	return nil
}

func getUintAndStringPtrToVersion(c *AzureSEVSNP, versionType attestationapi.Type, aux *fusedAzureSEVSNP) (*uint8, *string) {
	var v *uint8
	var stringV *string
	switch versionType {
	case attestationapi.Bootloader:
		v = &c.BootloaderVersion
		stringV = &aux.BootloaderVersion
	case attestationapi.TEE:
		v = &c.TEEVersion
		stringV = &aux.TEEVersion
	case attestationapi.SNP:
		v = &c.SNPVersion
		stringV = &aux.SNPVersion
	case attestationapi.Microcode:
		v = &c.MicrocodeVersion
		stringV = &aux.MicrocodeVersion
	}
	return v, stringV
}

// AzureTrustedLaunch is the configuration for Azure Trusted Launch attestation.
type AzureTrustedLaunch struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns azure-trusted-launch as the variant.
func (AzureTrustedLaunch) GetVariant() variant.Variant {
	return variant.AzureTrustedLaunch{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AzureTrustedLaunch) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AzureTrustedLaunch) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AzureTrustedLaunch) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AzureTrustedLaunch)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// auxAzureSEVSNP is a helper struct for unmarshaling the config from YAML for handling the version parsing.
// The version fields are kept to make it convertable to the native AzureSEVSNP struct.
type auxAzureSEVSNP struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// description: |
	//   Lowest acceptable bootloader version.
	BootloaderVersion uint8 `yaml:"-"`
	// description: |
	//   Lowest acceptable TEE version.
	TEEVersion uint8 `json:"teeVersion" yaml:"-"`
	// description: |
	//   Lowest acceptable SEV-SNP version.
	SNPVersion uint8 `json:"snpVersion" yaml:"-"`
	// description: |
	//   Lowest acceptable microcode version.
	MicrocodeVersion uint8 `json:"microcodeVersion" yaml:"-"`
	// description: |
	//   Configuration for validating the firmware signature.
	FirmwareSignerConfig SNPFirmwareSignerConfig `json:"firmwareSignerConfig" yaml:"firmwareSignerConfig"`
	// description: |
	//   AMD Root Key certificate used to verify the SEV-SNP certificate chain.
	AMDRootKey Certificate `json:"amdRootKey" yaml:"amdRootKey"`
}

// fusedAzureSEVSNP is a helper struct for unmarshaling the config from YAML for handling the version parsing.
type fusedAzureSEVSNP struct {
	*auxAzureSEVSNP `yaml:",inline"`
	// description: |
	//   Lowest acceptable bootloader version.
	BootloaderVersion string `yaml:"bootloaderVersion"`
	// description: |
	//   Lowest acceptable TEE version.
	TEEVersion string `yaml:"teeVersion"`
	// description: |
	//   Lowest acceptable SEV-SNP version.
	SNPVersion string `yaml:"snpVersion"`
	// description: |
	//   Lowest acceptable microcode version.
	MicrocodeVersion string `yaml:"microcodeVersion"`
}

func convertStringToUint(c *AzureSEVSNP, versionType attestationapi.Type, aux *fusedAzureSEVSNP) error {
	v, stringV := getUintAndStringPtrToVersion(c, versionType, aux)

	bvInt, err := strconv.ParseInt(*stringV, 10, 8)
	if err != nil {
		return err
	}
	*v = uint8(bvInt)
	return nil
}

func convertLatestToNumber(c *AzureSEVSNP, versions versionsapi.AzureSEVSNPVersion, versionType attestationapi.Type, aux *fusedAzureSEVSNP) bool {
	v, stringV := getUintAndStringPtrToVersion(c, versionType, aux)
	if strings.ToLower(*stringV) == "latest" {
		*v = attestationapi.GetVersionByType(versions, versionType)
		return true
	}
	return false
}
