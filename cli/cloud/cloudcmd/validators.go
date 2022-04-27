package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/config"
)

const warningStr = "Warning: not verifying the Constellation's %s measurements\n"

type Validators struct {
	provider   cloudprovider.Provider
	pcrs       map[uint32][]byte
	validators []atls.Validator
}

func NewValidators(provider cloudprovider.Provider, config *config.Config) (*Validators, error) {
	v := Validators{}
	if provider == cloudprovider.Unknown {
		return nil, errors.New("unknown cloud provider")
	}
	v.provider = provider
	if err := v.setPCRs(config); err != nil {
		return nil, err
	}
	return &v, nil
}

func (v *Validators) UpdateInitPCRs(ownerID, clusterID string) error {
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
func (v *Validators) updatePCR(pcrIndex uint32, encoded string) error {
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
	expectedPcr := sha256.Sum256(append(v.pcrs[pcrIndex], hashedInput[:]...))
	v.pcrs[pcrIndex] = expectedPcr[:]
	return nil
}

func (v *Validators) setPCRs(config *config.Config) error {
	switch v.provider {
	case cloudprovider.GCP:
		gcpPCRs := *config.Provider.GCP.PCRs
		if err := v.checkPCRs(gcpPCRs); err != nil {
			return err
		}
		v.pcrs = gcpPCRs
	case cloudprovider.Azure:
		azurePCRs := *config.Provider.Azure.PCRs
		if err := v.checkPCRs(azurePCRs); err != nil {
			return err
		}
		v.pcrs = azurePCRs
	}
	return nil
}

// V returns validators as list of atls.Validator.
func (v *Validators) V() []atls.Validator {
	v.updateValidators()
	return v.validators
}

func (v *Validators) updateValidators() {
	switch v.provider {
	case cloudprovider.GCP:
		v.validators = []atls.Validator{
			gcp.NewValidator(v.pcrs),
			gcp.NewNonCVMValidator(map[uint32][]byte{}), // TODO: Remove once we no longer use non CVMs.
		}
	case cloudprovider.Azure:
		v.validators = []atls.Validator{
			azure.NewValidator(v.pcrs),
		}
	}
}

// Warnings returns warnings for the specifc PCR values that are not verified.
//
// PCR allocation inspired by https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12#Tab1
func (v *Validators) Warnings() string {
	sb := &strings.Builder{}

	if v.pcrs[0] == nil || v.pcrs[1] == nil {
		writeFmt(sb, warningStr, "BIOS")
	}

	if v.pcrs[2] == nil || v.pcrs[3] == nil {
		writeFmt(sb, warningStr, "OPROM")
	}

	if v.pcrs[4] == nil || v.pcrs[5] == nil {
		writeFmt(sb, warningStr, "MBR")
	}

	// GRUB measures kernel command line and initrd into pcrs 8 and 9
	if v.pcrs[8] == nil {
		writeFmt(sb, warningStr, "kernel command line")
	}
	if v.pcrs[9] == nil {
		writeFmt(sb, warningStr, "initrd")
	}

	return sb.String()
}

// WarningsIncludeInit returns warnings for the specifc PCR values that are not verified.
// Warnings regarding the initialization are included.
//
// PCR allocation inspired by https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12#Tab1
func (v *Validators) WarningsIncludeInit() string {
	warnings := v.Warnings()
	if v.pcrs[uint32(vtpm.PCRIndexOwnerID)] == nil || v.pcrs[uint32(vtpm.PCRIndexClusterID)] == nil {
		warnings = warnings + fmt.Sprintf(warningStr, "initialization status")
	}

	return warnings
}

func (v *Validators) checkPCRs(pcrs map[uint32][]byte) error {
	if len(pcrs) == 0 {
		return errors.New("no PCR values provided")
	}
	for k, v := range pcrs {
		if len(v) != 32 {
			return fmt.Errorf("bad config: PCR[%d]: expected length: %d, but got: %d", k, 32, len(v))
		}
	}
	return nil
}

func writeFmt(sb *strings.Builder, fmtStr string, args ...interface{}) {
	sb.WriteString(fmt.Sprintf(fmtStr, args...))
}
