/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package config

import (
	"bytes"
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/encoding"
)

var _ sevsnpMarshaller = &AzureSEVSNP{}

// DefaultForAzureSEVSNP returns the default configuration for Azure SEV-SNP attestation.
// Version numbers have placeholder values and the latest available values can be fetched using [AzureSEVSNP.FetchAndSetLatestVersionNumbers].
func DefaultForAzureSEVSNP() *AzureSEVSNP {
	return &AzureSEVSNP{
		Measurements:      measurements.DefaultsFor(cloudprovider.Azure, variant.AzureSEVSNP{}),
		BootloaderVersion: NewLatestPlaceholderVersion(),
		TEEVersion:        NewLatestPlaceholderVersion(),
		SNPVersion:        NewLatestPlaceholderVersion(),
		MicrocodeVersion:  NewLatestPlaceholderVersion(),
		FirmwareSignerConfig: SNPFirmwareSignerConfig{
			AcceptedKeyDigests: idkeydigest.DefaultList(),
			EnforcementPolicy:  idkeydigest.MAAFallback,
		},
		AMDRootKey: mustParsePEM(arkPEM),
	}
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

// FetchAndSetLatestVersionNumbers fetches the latest version numbers from the configapi and sets them.
func (c *AzureSEVSNP) FetchAndSetLatestVersionNumbers(ctx context.Context, fetcher attestationconfigapi.Fetcher) error {
	// Only talk to the API if at least one version number is set to latest.
	if !(c.BootloaderVersion.WantLatest || c.TEEVersion.WantLatest || c.SNPVersion.WantLatest || c.MicrocodeVersion.WantLatest) {
		return nil
	}

	versions, err := fetcher.FetchSEVSNPVersionLatest(ctx, variant.AzureSEVSNP{})
	if err != nil {
		return fmt.Errorf("fetching latest TCB versions from configapi: %w", err)
	}
	// set number and keep isLatest flag
	c.mergeWithLatestVersion(versions.SEVSNPVersion)
	return nil
}

func (c *AzureSEVSNP) mergeWithLatestVersion(latest attestationconfigapi.SEVSNPVersion) {
	if c.BootloaderVersion.WantLatest {
		c.BootloaderVersion.Value = latest.Bootloader
	}
	if c.TEEVersion.WantLatest {
		c.TEEVersion.Value = latest.TEE
	}
	if c.SNPVersion.WantLatest {
		c.SNPVersion.Value = latest.SNP
	}
	if c.MicrocodeVersion.WantLatest {
		c.MicrocodeVersion.Value = latest.Microcode
	}
}

func (c *AzureSEVSNP) getToMarshallLatestWithResolvedVersions() AttestationCfg {
	cp := *c
	cp.BootloaderVersion.WantLatest = false
	cp.TEEVersion.WantLatest = false
	cp.SNPVersion.WantLatest = false
	cp.MicrocodeVersion.WantLatest = false
	return &cp
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

// DefaultForAzureTDX returns the default configuration for Azure TDX attestation.
func DefaultForAzureTDX() *AzureTDX {
	return &AzureTDX{
		Measurements: measurements.DefaultsFor(cloudprovider.Azure, variant.AzureTDX{}),
		// TODO: Set default values for version numbers once enabled.
		QESVN:      0,
		PCESVN:     0,
		TEETCBSVN:  encoding.HexBytes(bytes.Repeat([]byte{0x00}, 16)), // equivalent of accepting all TEE versions
		QEVendorID: encoding.HexBytes(bytes.Repeat([]byte{0x00}, 16)), // TODO: Decide on consistent value or remove
		MRSeam:     encoding.HexBytes(bytes.Repeat([]byte{0x00}, 48)), // TODO: Decide on consistent value or remove
		XFAM:       encoding.HexBytes(bytes.Repeat([]byte{0x00}, 8)),  // TODO: Decide on consistent value or remove

		IntelRootKey: mustParsePEM(tdxRootPEM),
	}
}

// GetVariant returns azure-tdx as the variant.
func (AzureTDX) GetVariant() variant.Variant {
	return variant.AzureTDX{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AzureTDX) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AzureTDX) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AzureTDX) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AzureTDX)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

func (c *AzureTDX) getToMarshallLatestWithResolvedVersions() AttestationCfg {
	cp := *c
	// TODO: We probably want to support "latest" pseudo versioning for Azure TDX
	// But we should decide on which claims can be reliably used for attestation first
	return &cp
}
