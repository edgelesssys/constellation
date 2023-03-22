/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/spf13/cobra"
)

// NewValidator creates a new Validator.
func NewValidator(cmd *cobra.Command, config config.AttestationCfg, log debugLog) (atls.Validator, error) {
	return choose.Validator(config, warnLogger{cmd: cmd, log: log})
}

// UpdateInitMeasurements sets the owner and cluster measurement values.
func UpdateInitMeasurements(config config.AttestationCfg, ownerID, clusterID string) error {
	m := config.GetMeasurements()
	if err := updateMeasurement(m, uint32(measurements.PCRIndexOwnerID), ownerID); err != nil {
		return err
	}
	return updateMeasurement(m, uint32(measurements.PCRIndexClusterID), clusterID)
}

// updateMeasurement adds a new entry to the measurements of v, or removes the key if the input is an empty string.
//
// When adding, the input is first decoded from hex or base64.
// We then calculate the expected measurement by hashing the input using SHA256,
// appending expected measurement for initialization, and then hashing once more.
func updateMeasurement(m measurements.M, measurementIdx uint32, encoded string) error {
	if encoded == "" {
		delete(m, measurementIdx)
		return nil
	}

	// decode from hex or base64
	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		hexErr := err
		decoded, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return fmt.Errorf("input [%s] could neither be hex decoded (%w) nor base64 decoded (%w)", encoded, hexErr, err)
		}
	}
	// new_measurement_value := hash(old_pcr_value || data_to_extend)
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

// warnLogger implements logging of warnings for validators.
type warnLogger struct {
	cmd *cobra.Command
	log debugLog
}

// Infof messages are reduced to debug messages, since we don't want
// the extra info when using the CLI without setting the debug flag.
func (wl warnLogger) Infof(fmtStr string, args ...any) {
	wl.log.Debugf(fmtStr, args...)
}

// Warnf prints a formatted warning from the validator.
func (wl warnLogger) Warnf(fmtStr string, args ...any) {
	wl.cmd.PrintErrf("Warning: %s\n", fmt.Sprintf(fmtStr, args...))
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}
