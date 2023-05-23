/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// QEMUVTPM is the configuration for QEMU vTPM attestation.
type QEMUVTPM struct {
	// description: |
	//   Expected TPM measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns qemu-vtpm as the variant.
func (QEMUVTPM) GetVariant() variant.Variant {
	return variant.QEMUVTPM{}
}

// GetMeasurements returns the measurements used for attestation.
func (c QEMUVTPM) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *QEMUVTPM) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c QEMUVTPM) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*QEMUVTPM)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}

// QEMUTDX is the configuration for QEMU TDX attestation.
type QEMUTDX struct {
	// description: |
	//   Expected TDX measurements.
	Measurements measurements.M `json:"measurements" yaml:"measurements" validate:"required,no_placeholders"`
}

// GetVariant returns qemu-tdx as the variant.
func (QEMUTDX) GetVariant() variant.Variant {
	return variant.QEMUTDX{}
}

// GetMeasurements returns the measurements used for attestation.
func (c QEMUTDX) GetMeasurements() measurements.M {
	return c.Measurements
}

// SetMeasurements updates a config's measurements using the given measurements.
func (c *QEMUTDX) SetMeasurements(m measurements.M) {
	c.Measurements = m
}

// EqualTo returns true if the config is equal to the given config.
func (c QEMUTDX) EqualTo(other AttestationCfg) (bool, error) {
	otherCfg, ok := other.(*QEMUTDX)
	if !ok {
		return false, fmt.Errorf("cannot compare %T with %T", c, other)
	}
	return c.Measurements.EqualTo(otherCfg.Measurements), nil
}
