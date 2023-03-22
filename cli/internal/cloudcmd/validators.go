/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/spf13/cobra"
)

// Validator validates Measurements (Platform Control Registers = PCRs).
type Validator struct {
	attestationVariant oid.Getter
	measurements       measurements.M
	idkeydigests       idkeydigest.IDKeyDigests
	enforceIDKeyDigest bool
	validator          atls.Validator
	log                debugLog
}

// NewValidator creates a new Validator.
func NewValidator(conf *config.Config, log debugLog) (*Validator, error) {
	v := Validator{log: log}
	variant, err := oid.FromString(conf.AttestationVariant)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation variant: %w", err)
	}
	v.attestationVariant = variant // valid variant

	if err := v.setMeasurements(conf); err != nil {
		return nil, err
	}

	if v.attestationVariant.OID().Equal(oid.AzureSEVSNP{}.OID()) {
		v.enforceIDKeyDigest = conf.EnforcesIDKeyDigest()
		v.idkeydigests = conf.IDKeyDigests()
	}

	return &v, nil
}

// UpdateInitMeasurements sets the owner and cluster measurement values.
func (v *Validator) UpdateInitMeasurements(ownerID, clusterID string) error {
	if err := v.updateMeasurement(uint32(measurements.PCRIndexOwnerID), ownerID); err != nil {
		return err
	}
	return v.updateMeasurement(uint32(measurements.PCRIndexClusterID), clusterID)
}

// updateMeasurement adds a new entry to the measurements of v, or removes the key if the input is an empty string.
//
// When adding, the input is first decoded from hex or base64.
// We then calculate the expected measurement by hashing the input using SHA256,
// appending expected measurement for initialization, and then hashing once more.
func (v *Validator) updateMeasurement(measurementIndex uint32, encoded string) error {
	if encoded == "" {
		delete(v.measurements, measurementIndex)
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
	oldExpected := v.measurements[measurementIndex].Expected
	expectedMeasurement := sha256.Sum256(append(oldExpected[:], hashedInput[:]...))
	v.measurements[measurementIndex] = measurements.Measurement{
		Expected: expectedMeasurement[:],
		WarnOnly: v.measurements[measurementIndex].WarnOnly,
	}
	return nil
}

func (v *Validator) setMeasurements(config *config.Config) error {
	switch v.attestationVariant {
	case oid.AWSNitroTPM{}:
		awsMeasurements := config.Provider.AWS.Measurements
		if len(awsMeasurements) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.measurements = awsMeasurements
	case oid.AzureSEVSNP{}, oid.AzureTrustedLaunch{}:
		azureMeasurements := config.Provider.Azure.Measurements
		if len(azureMeasurements) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.measurements = azureMeasurements
	case oid.GCPSEVES{}:
		gcpMeasurements := config.Provider.GCP.Measurements
		if len(gcpMeasurements) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.measurements = gcpMeasurements
	case oid.QEMUVTPM{}, oid.QEMUTDX{}:
		qemuMeasurements := config.Provider.QEMU.Measurements
		if len(qemuMeasurements) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.measurements = qemuMeasurements
	}
	return nil
}

// V returns the validator as atls.Validator.
func (v *Validator) V(cmd *cobra.Command) atls.Validator {
	v.updateValidator(cmd)
	return v.validator
}

// Measurements returns the validator's measurements map.
func (v *Validator) Measurements() measurements.M {
	return v.measurements
}

func (v *Validator) updateValidator(cmd *cobra.Command) {
	log := warnLogger{cmd: cmd, log: v.log}

	// Use of a valid variant has been check in NewValidator so we may drop the error
	v.validator, _ = choose.Validator(v.attestationVariant, v.measurements, v.idkeydigests, v.enforceIDKeyDigest, log)
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
