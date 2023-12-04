/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/spf13/cobra"
)

// NewValidator creates a new Validator.
func NewValidator(cmd *cobra.Command, config config.AttestationCfg, log debugLog) (atls.Validator, error) {
	return choose.Validator(config, WarnLogger{Cmd: cmd, Log: log})
}

// UpdateInitMeasurements sets the owner and cluster measurement values.
func UpdateInitMeasurements(config config.AttestationCfg, ownerID, clusterID string) error {
	m := config.GetMeasurements()

	switch config.GetVariant() {
	case variant.AWSNitroTPM{}, variant.AWSSEVSNP{}, variant.AzureTrustedLaunch{}, variant.AzureSEVSNP{}, variant.GCPSEVES{}, variant.QEMUVTPM{}:
		if err := updateMeasurementTPM(m, uint32(measurements.PCRIndexOwnerID), ownerID); err != nil {
			return err
		}
		return updateMeasurementTPM(m, uint32(measurements.PCRIndexClusterID), clusterID)
	case variant.QEMUTDX{}:
		// Measuring ownerID is currently not implemented for Constellation
		// Since adding support for measuring ownerID to TDX would require additional code changes,
		// the current implementation does not support it, but can be changed if we decide to add support in the future
		return updateMeasurementTDX(m, uint32(measurements.TDXIndexClusterID), clusterID)
	default:
		return fmt.Errorf("selecting attestation variant: unknown attestation variant")
	}
}

func updateMeasurementTDX(m measurements.M, measurementIdx uint32, encoded string) error {
	if encoded == "" {
		delete(m, measurementIdx)
		return nil
	}
	decoded, err := decodeMeasurement(encoded)
	if err != nil {
		return err
	}

	// new_measurement_value := hash(old_measurement_value || data_to_extend)
	// Since we use the DG.MR.RTMR.EXTEND call to extend the register, data_to_extend is the hash of our input
	hashedInput := sha512.Sum384(decoded)
	oldExpected := m[measurementIdx].Expected
	expectedMeasurementSum := sha512.Sum384(append(oldExpected[:], hashedInput[:]...))
	m[measurementIdx] = measurements.Measurement{
		Expected:      expectedMeasurementSum[:],
		ValidationOpt: m[measurementIdx].ValidationOpt,
	}
	return nil
}

func updateMeasurementTPM(m measurements.M, measurementIdx uint32, encoded string) error {
	if encoded == "" {
		delete(m, measurementIdx)
		return nil
	}
	decoded, err := decodeMeasurement(encoded)
	if err != nil {
		return err
	}

	// new_pcr_value := hash(old_pcr_value || data_to_extend)
	// Since we use the TPM2_PCR_Event call to extend the PCR, data_to_extend is the hash of our input
	hashedInput := sha256.Sum256(decoded)
	oldExpected := m[measurementIdx].Expected
	expectedMeasurement := sha256.Sum256(append(oldExpected[:], hashedInput[:]...))
	m[measurementIdx] = measurements.Measurement{
		Expected:      expectedMeasurement[:],
		ValidationOpt: m[measurementIdx].ValidationOpt,
	}
	return nil
}

func decodeMeasurement(encoded string) ([]byte, error) {
	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		hexErr := err
		decoded, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("input [%s] could neither be hex decoded (%w) nor base64 decoded (%w)", encoded, hexErr, err)
		}
	}
	return decoded, nil
}

// WarnLogger implements logging of warnings for validators.
type WarnLogger struct {
	Cmd *cobra.Command
	Log debugLog
}

// Infof messages are reduced to debug messages, since we don't want
// the extra info when using the CLI without setting the debug flag.
func (wl WarnLogger) Infof(fmtStr string, args ...any) {
	wl.Log.Debugf(fmtStr, args...)
}

// Warnf prints a formatted warning from the validator.
func (wl WarnLogger) Warnf(fmtStr string, args ...any) {
	wl.Cmd.PrintErrf("Warning: %s\n", fmt.Sprintf(fmtStr, args...))
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}
