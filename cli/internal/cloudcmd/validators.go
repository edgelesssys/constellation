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
	"github.com/edgelesssys/constellation/v2/internal/attestation/aws"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/spf13/cobra"
)

// Validator validates Platform Configuration Registers (PCRs).
type Validator struct {
	provider           cloudprovider.Provider
	pcrs               measurements.M
	idkeydigest        []byte
	enforceIDKeyDigest bool
	azureCVM           bool
	validator          atls.Validator
}

// NewValidator creates a new Validator.
func NewValidator(provider cloudprovider.Provider, conf *config.Config) (*Validator, error) {
	v := Validator{}
	if provider == cloudprovider.Unknown {
		return nil, errors.New("unknown cloud provider")
	}
	v.provider = provider
	if err := v.setPCRs(conf); err != nil {
		return nil, err
	}

	if v.provider == cloudprovider.Azure {
		v.azureCVM = *conf.Provider.Azure.ConfidentialVM
		if v.azureCVM {
			idkeydigest, err := hex.DecodeString(conf.Provider.Azure.IDKeyDigest)
			if err != nil {
				return nil, fmt.Errorf("bad config: decoding idkeydigest from config: %w", err)
			}
			v.enforceIDKeyDigest = *conf.Provider.Azure.EnforceIDKeyDigest
			v.idkeydigest = idkeydigest
		}
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

// updatePCR adds a new entry to the pcr map of v, or removes the key if the input is an empty string.
//
// When adding, the input is first decoded from base64.
// We then calculate the expected PCR by hashing the input using SHA256,
// appending expected PCR for initialization, and then hashing once more.
func (v *Validator) updatePCR(pcrIndex uint32, encoded string) error {
	if encoded == "" {
		delete(v.pcrs, pcrIndex)
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("input [%s] is not base64 encoded: %w", encoded, err)
	}
	// new_pcr_value := hash(old_pcr_value || data_to_extend)
	// Since we use the TPM2_PCR_Event call to extend the PCR, data_to_extend is the hash of our input
	hashedInput := sha256.Sum256(decoded)
	oldExpected := v.pcrs[pcrIndex].Expected
	expectedPcr := sha256.Sum256(append(oldExpected[:], hashedInput[:]...))
	v.pcrs[pcrIndex] = measurements.Measurement{
		Expected: expectedPcr,
		WarnOnly: v.pcrs[pcrIndex].WarnOnly,
	}
	return nil
}

func (v *Validator) setPCRs(config *config.Config) error {
	switch v.provider {
	case cloudprovider.AWS:
		awsPCRs := config.Provider.AWS.Measurements
		if len(awsPCRs) == 0 {
			return errors.New("no PCR values provided")
		}
		v.pcrs = awsPCRs
	case cloudprovider.Azure:
		azurePCRs := config.Provider.Azure.Measurements
		if len(azurePCRs) == 0 {
			return errors.New("no PCR values provided")
		}
		v.pcrs = azurePCRs
	case cloudprovider.GCP:
		gcpPCRs := config.Provider.GCP.Measurements
		if len(gcpPCRs) == 0 {
			return errors.New("no PCR values provided")
		}
		v.pcrs = gcpPCRs
	case cloudprovider.QEMU:
		qemuPCRs := config.Provider.QEMU.Measurements
		if len(qemuPCRs) == 0 {
			return errors.New("no PCR values provided")
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
	log := warnLogger{cmd: cmd}
	switch v.provider {
	case cloudprovider.GCP:
		v.validator = gcp.NewValidator(v.pcrs, log)
	case cloudprovider.Azure:
		if v.azureCVM {
			v.validator = snp.NewValidator(v.pcrs, v.idkeydigest, v.enforceIDKeyDigest, log)
		} else {
			v.validator = trustedlaunch.NewValidator(v.pcrs, log)
		}
	case cloudprovider.AWS:
		v.validator = aws.NewValidator(v.pcrs, log)
	case cloudprovider.QEMU:
		v.validator = qemu.NewValidator(v.pcrs, log)
	}
}

// warnLogger implements logging of warnings for validators.
type warnLogger struct {
	cmd *cobra.Command
}

// Infof is a no-op since we don't want extra info messages when using the CLI.
func (wl warnLogger) Infof(format string, args ...any) {}

// Warnf prints a formatted warning from the validator.
func (wl warnLogger) Warnf(fmtStr string, args ...any) {
	wl.cmd.PrintErrf("Warning: %s\n", fmt.Sprintf(fmtStr, args...))
}
