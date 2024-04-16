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
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

var _ svnResolveMarshaller = &GCPSEVSNP{}

// DefaultForGCPSEVSNP provides a valid default configuration for GCP SEV-SNP attestation.
func DefaultForGCPSEVSNP() *GCPSEVSNP {
	return &GCPSEVSNP{
		Measurements:      measurements.DefaultsFor(cloudprovider.GCP, variant.GCPSEVSNP{}),
		BootloaderVersion: NewLatestPlaceholderVersion(),
		TEEVersion:        NewLatestPlaceholderVersion(),
		SNPVersion:        NewLatestPlaceholderVersion(),
		MicrocodeVersion:  NewLatestPlaceholderVersion(),
		AMDRootKey:        mustParsePEM(arkPEM),
	}
}

// GetVariant returns gcp-sev-snp as the variant.
func (GCPSEVSNP) GetVariant() variant.Variant {
	return variant.GCPSEVSNP{}
}

// GetMeasurements returns the measurements used for attestation.
func (c GCPSEVSNP) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *GCPSEVSNP) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c GCPSEVSNP) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*GCPSEVSNP)
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

func (c *GCPSEVSNP) getToMarshallLatestWithResolvedVersions() AttestationCfg {
	cp := *c
	cp.BootloaderVersion.WantLatest = false
	cp.TEEVersion.WantLatest = false
	cp.SNPVersion.WantLatest = false
	cp.MicrocodeVersion.WantLatest = false
	return &cp
}

// FetchAndSetLatestVersionNumbers fetches the latest version numbers from the configapi and sets them.
func (c *GCPSEVSNP) FetchAndSetLatestVersionNumbers(ctx context.Context, fetcher attestationconfigapi.Fetcher) error {
	// Only talk to the API if at least one version number is set to latest.
	if !(c.BootloaderVersion.WantLatest || c.TEEVersion.WantLatest || c.SNPVersion.WantLatest || c.MicrocodeVersion.WantLatest) {
		return nil
	}

	versions, err := fetcher.FetchSEVSNPVersionLatest(ctx, variant.GCPSEVSNP{})
	if err != nil {
		return fmt.Errorf("fetching latest TCB versions from configapi: %w", err)
	}
	// set number and keep isLatest flag
	c.mergeWithLatestVersion(versions.SEVSNPVersion)
	return nil
}

func (c *GCPSEVSNP) mergeWithLatestVersion(latest attestationconfigapi.SEVSNPVersion) {
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

// GetVariant returns gcp-sev-es as the variant.
func (GCPSEVES) GetVariant() variant.Variant {
	return variant.GCPSEVES{}
}

// GetMeasurements returns the measurements used for attestation.
func (c GCPSEVES) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *GCPSEVES) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c GCPSEVES) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*GCPSEVES)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}
