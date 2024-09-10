/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package config

import (
	"bytes"
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

var _ svnResolveMarshaller = &AWSSEVSNP{}

// DefaultForAWSSEVSNP provides a valid default configuration for AWS SEV-SNP attestation.
func DefaultForAWSSEVSNP() *AWSSEVSNP {
	return &AWSSEVSNP{
		Measurements:      measurements.DefaultsFor(cloudprovider.AWS, variant.AWSSEVSNP{}),
		BootloaderVersion: NewLatestPlaceholderVersion[uint8](),
		TEEVersion:        NewLatestPlaceholderVersion[uint8](),
		SNPVersion:        NewLatestPlaceholderVersion[uint8](),
		MicrocodeVersion:  NewLatestPlaceholderVersion[uint8](),
		AMDRootKey:        mustParsePEM(arkPEM),
	}
}

// GetVariant returns aws-sev-snp as the variant.
func (AWSSEVSNP) GetVariant() variant.Variant {
	return variant.AWSSEVSNP{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AWSSEVSNP) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AWSSEVSNP) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AWSSEVSNP) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AWSSEVSNP)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}

	measurementsEqual := c.Measurements.EqualTo(otherCfg.Measurements)
	bootloaderEqual := c.BootloaderVersion == otherCfg.BootloaderVersion
	teeEqual := c.TEEVersion == otherCfg.TEEVersion
	snpEqual := c.SNPVersion == otherCfg.SNPVersion
	microcodeEqual := c.MicrocodeVersion == otherCfg.MicrocodeVersion
	rootKeyEqual := bytes.Equal(c.AMDRootKey.Raw, otherCfg.AMDRootKey.Raw)
	signingKeyEqual := bytes.Equal(c.AMDSigningKey.Raw, otherCfg.AMDSigningKey.Raw)

	return measurementsEqual && bootloaderEqual && teeEqual && snpEqual && microcodeEqual && rootKeyEqual && signingKeyEqual, nil
}

func (c *AWSSEVSNP) getToMarshallLatestWithResolvedVersions() AttestationCfg {
	cp := *c
	cp.BootloaderVersion.WantLatest = false
	cp.TEEVersion.WantLatest = false
	cp.SNPVersion.WantLatest = false
	cp.MicrocodeVersion.WantLatest = false
	return &cp
}

// FetchAndSetLatestVersionNumbers fetches the latest version numbers from the configapi and sets them.
func (c *AWSSEVSNP) FetchAndSetLatestVersionNumbers(ctx context.Context, fetcher attestationconfig.Fetcher) error {
	// Only talk to the API if at least one version number is set to latest.
	if !(c.BootloaderVersion.WantLatest || c.TEEVersion.WantLatest || c.SNPVersion.WantLatest || c.MicrocodeVersion.WantLatest) {
		return nil
	}

	versions, err := fetcher.FetchLatestVersion(ctx, variant.AWSSEVSNP{})
	if err != nil {
		return fmt.Errorf("fetching latest TCB versions from configapi: %w", err)
	}
	// set number and keep isLatest flag
	c.mergeWithLatestVersion(versions.SEVSNPVersion)
	return nil
}

func (c *AWSSEVSNP) mergeWithLatestVersion(latest attestationconfig.SEVSNPVersion) {
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

// GetVariant returns aws-nitro-tpm as the variant.
func (AWSNitroTPM) GetVariant() variant.Variant {
	return variant.AWSNitroTPM{}
}

// GetMeasurements returns the measurements used for attestation.
func (c AWSNitroTPM) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *AWSNitroTPM) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c AWSNitroTPM) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*AWSNitroTPM)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}
