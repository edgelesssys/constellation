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

// Validator validates Platform Configuration Registers (PCRs).
type Validator struct {
	attestationVariant oid.Getter
	pcrs               measurements.M
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

	if err := v.setPCRs(conf); err != nil {
		return nil, err
	}

	if v.attestationVariant.OID().Equal(oid.AzureSEVSNP{}.OID()) {
		v.enforceIDKeyDigest = conf.EnforcesIDKeyDigest()
		v.idkeydigests = conf.IDKeyDigests()
	}

	return &v, nil
}

// UpdateInitPCRs sets the owner and cluster PCR values.
func (v *Validator) UpdateInitPCRs(ownerID, clusterID string) error {
	if err := v.updatePCR(uint32(measurements.PCRIndexOwnerID), ownerID); err != nil {
		return err
	}
	return v.updatePCR(uint32(measurements.PCRIndexClusterID), clusterID)
}

// updatePCR adds a new entry to the measurements of v, or removes the key if the input is an empty string.
//
// When adding, the input is first decoded from hex or base64.
// We then calculate the expected PCR by hashing the input using SHA256,
// appending expected PCR for initialization, and then hashing once more.
func (v *Validator) updatePCR(pcrIndex uint32, encoded string) error {
	if encoded == "" {
		delete(v.pcrs, pcrIndex)
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
	// new_pcr_value := hash(old_pcr_value || data_to_extend)
	// Since we use the TPM2_PCR_Event call to extend the PCR, data_to_extend is the hash of our input
	hashedInput := sha256.Sum256(decoded)
	oldExpected := v.pcrs[pcrIndex].Expected
	expectedPcr := sha256.Sum256(append(oldExpected[:], hashedInput[:]...))
	v.pcrs[pcrIndex] = measurements.Measurement{
		Expected: expectedPcr[:],
		WarnOnly: v.pcrs[pcrIndex].WarnOnly,
	}
	return nil
}

func (v *Validator) setPCRs(config *config.Config) error {
	switch v.attestationVariant {
	case oid.AWSNitroTPM{}:
		awsPCRs := config.Provider.AWS.Measurements
		if len(awsPCRs) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.pcrs = awsPCRs
	case oid.AzureSEVSNP{}, oid.AzureTrustedLaunch{}:
		azurePCRs := config.Provider.Azure.Measurements
		if len(azurePCRs) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.pcrs = azurePCRs
	case oid.GCPSEVES{}:
		gcpPCRs := config.Provider.GCP.Measurements
		if len(gcpPCRs) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.pcrs = gcpPCRs
	case oid.QEMUVTPM{}:
		qemuPCRs := config.Provider.QEMU.Measurements
		if len(qemuPCRs) == 0 {
			return errors.New("no expected measurement provided")
		}
		v.pcrs = qemuPCRs
	}
	return nil
}

// V returns the validator as atls.Validator.
func (v *Validator) V(cmd *cobra.Command) atls.Validator {
	v.updateValidator(cmd)
	return v.validator
}

// PCRS returns the validator's PCR map.
func (v *Validator) PCRS() measurements.M {
	return v.pcrs
}

func (v *Validator) updateValidator(cmd *cobra.Command) {
	log := warnLogger{cmd: cmd, log: v.log}

	// Use of a valid variant has been check in NewValidator so we may drop the error
	v.validator, _ = choose.Validator(v.attestationVariant, v.pcrs, v.idkeydigests, v.enforceIDKeyDigest, log)
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
