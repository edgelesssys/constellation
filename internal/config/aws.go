/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package config

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/sev-snp-measure-go/ovmf"
)

// AWSSEVSNP is the configuration for AWS SEV-SNP attestation.
type AWSSEVSNP struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
	// TODO (derpsteb): reenable launchMeasurement once we have a way to generate the expected value dynamically.
	// description: |
	//   Expected launch measurement in SNP report. Not in use right now.
	OVMFFirmware OVMFConfig
	// description: |
	//   AMD Root Key certificate used to verify the SEV-SNP certificate chain.
	AMDRootKey Certificate `json:"amdRootKey" yaml:"amdRootKey"`
}

type OVMFConfig struct {
	WantLatest bool
	Metadata   []ovmf.MetadataWrapper `json:"metadata"`
}

// DefaultForAWSSEVSNP provides a valid default configuration for AWS SEV-SNP attestation.
func DefaultForAWSSEVSNP() *AWSSEVSNP {
	return &AWSSEVSNP{
		Measurements: measurements.DefaultsFor(cloudprovider.AWS, variant.AWSSEVSNP{}),
		// LaunchMeasurement: measurements.PlaceHolderMeasurement(48),
		AMDRootKey: mustParsePEM(constants.AMDRootKey),
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
	// TODO (derpsteb): reenable launchMeasurement once we have a way to generate the expected value dynamically.
	// if !bytes.Equal(c.LaunchMeasurement.Expected, otherCfg.LaunchMeasurement.Expected) {
	// 	return false, nil
	// }
	// if c.LaunchMeasurement.ValidationOpt != otherCfg.LaunchMeasurement.ValidationOpt {
	// 	return false, nil
	// }

	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// AWSNitroTPM is the configuration for AWS Nitro TPM attestation.
type AWSNitroTPM struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
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
