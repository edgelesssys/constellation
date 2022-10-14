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
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/spf13/cobra"
)

type Validator struct {
	provider           cloudprovider.Provider
	pcrs               map[uint32][]byte
	enforcedPCRs       []uint32
	idkeydigest        []byte
	enforceIDKeyDigest bool
	azureCVM           bool
	validator          atls.Validator
}

func NewValidator(provider cloudprovider.Provider, config *config.Config) (*Validator, error) {
	v := Validator{}
	if provider == cloudprovider.Unknown {
		return nil, errors.New("unknown cloud provider")
	}
	v.provider = provider
	if err := v.setPCRs(config); err != nil {
		return nil, err
	}

	if v.provider == cloudprovider.Azure {
		v.azureCVM = *config.Provider.Azure.ConfidentialVM
		if v.azureCVM {
			idkeydigest, err := hex.DecodeString(config.Provider.Azure.IDKeyDigest)
			if err != nil {
				return nil, fmt.Errorf("bad config: decoding idkeydigest from config: %w", err)
			}
			v.enforceIDKeyDigest = *config.Provider.Azure.EnforceIDKeyDigest
			v.idkeydigest = idkeydigest
		}
	}

	return &v, nil
}

func (v *Validator) UpdateInitPCRs(ownerID, clusterID string) error {
	if err := v.updatePCR(uint32(vtpm.PCRIndexOwnerID), ownerID); err != nil {
		return err
	}
	return v.updatePCR(uint32(vtpm.PCRIndexClusterID), clusterID)
}

// updatePCR adds a new entry to the pcr map of v, or removes the key if the input is an empty string.
//
// When adding, the input is first decoded from base64.
// We then calculate the expected PCR by hashing the input using SHA256,
// appending expected PCR for initialization, and then hashing once more.
func (v *Validator) updatePCR(pcrIndex uint32, encoded string) error {
	if encoded == "" {
		delete(v.pcrs, pcrIndex)

		// remove enforced PCR if it exists
		for i, enforcedIdx := range v.enforcedPCRs {
			if enforcedIdx == pcrIndex {
				v.enforcedPCRs[i] = v.enforcedPCRs[len(v.enforcedPCRs)-1]
				v.enforcedPCRs = v.enforcedPCRs[:len(v.enforcedPCRs)-1]
				break
			}
		}

		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("input [%s] is not base64 encoded: %w", encoded, err)
	}
	// new_pcr_value := hash(old_pcr_value || data_to_extend)
	// Since we use the TPM2_PCR_Event call to extend the PCR, data_to_extend is the hash of our input
	hashedInput := sha256.Sum256(decoded)
	expectedPcr := sha256.Sum256(append(v.pcrs[pcrIndex], hashedInput[:]...))
	v.pcrs[pcrIndex] = expectedPcr[:]
	return nil
}

func (v *Validator) setPCRs(config *config.Config) error {
	switch v.provider {
	case cloudprovider.GCP:
		gcpPCRs := config.Provider.GCP.Measurements
		enforcedPCRs := config.Provider.GCP.EnforcedMeasurements
		if err := v.checkPCRs(gcpPCRs, enforcedPCRs); err != nil {
			return err
		}
		v.enforcedPCRs = enforcedPCRs
		v.pcrs = gcpPCRs
	case cloudprovider.Azure:
		azurePCRs := config.Provider.Azure.Measurements
		enforcedPCRs := config.Provider.Azure.EnforcedMeasurements
		if err := v.checkPCRs(azurePCRs, enforcedPCRs); err != nil {
			return err
		}
		v.enforcedPCRs = enforcedPCRs
		v.pcrs = azurePCRs
	case cloudprovider.QEMU:
		qemuPCRs := config.Provider.QEMU.Measurements
		enforcedPCRs := config.Provider.QEMU.EnforcedMeasurements
		if err := v.checkPCRs(qemuPCRs, enforcedPCRs); err != nil {
			return err
		}
		v.enforcedPCRs = enforcedPCRs
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
func (v *Validator) PCRS() map[uint32][]byte {
	return v.pcrs
}

func (v *Validator) updateValidator(cmd *cobra.Command) {
	log := warnLogger{cmd: cmd}
	switch v.provider {
	case cloudprovider.GCP:
		v.validator = gcp.NewValidator(v.pcrs, v.enforcedPCRs, log)
	case cloudprovider.Azure:
		if v.azureCVM {
			v.validator = snp.NewValidator(v.pcrs, v.enforcedPCRs, v.idkeydigest, v.enforceIDKeyDigest, log)
		} else {
			v.validator = trustedlaunch.NewValidator(v.pcrs, v.enforcedPCRs, log)
		}
	case cloudprovider.QEMU:
		v.validator = qemu.NewValidator(v.pcrs, v.enforcedPCRs, log)
	}
}

func (v *Validator) checkPCRs(pcrs map[uint32][]byte, enforcedPCRs []uint32) error {
	if len(pcrs) == 0 {
		return errors.New("no PCR values provided")
	}
	for k, v := range pcrs {
		if len(v) != 32 {
			return fmt.Errorf("bad config: PCR[%d]: expected length: %d, but got: %d", k, 32, len(v))
		}
	}
	for _, v := range enforcedPCRs {
		if _, ok := pcrs[v]; !ok {
			return fmt.Errorf("bad config: PCR[%d] is enforced, but no expected measurement is provided", v)
		}
	}
	return nil
}

// warnLogger implements logging of warnings for validators.
type warnLogger struct {
	cmd *cobra.Command
}

// Infof is a no-op since we don't want extra info messages when using the CLI.
func (wl warnLogger) Infof(format string, args ...interface{}) {}

// Warnf prints a formatted warning from the validator.
func (wl warnLogger) Warnf(fmtStr string, args ...interface{}) {
	wl.cmd.PrintErrf("Warning: %s\n", fmt.Sprintf(fmtStr, args...))
}
