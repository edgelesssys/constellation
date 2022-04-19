package cloudcmd

import (
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

type Validators struct {
	validators      []atls.Validator
	pcrWarnings     string
	pcrWarningsInit string
}

func NewValidators(provider cloudprovider.Provider, config *config.Config) (Validators, error) {
	v := Validators{}
	switch provider {
	case cloudprovider.GCP:
		gcpPCRs := *config.Provider.GCP.PCRs
		if err := v.checkPCRs(gcpPCRs); err != nil {
			return Validators{}, err
		}
		v.setPCRWarnings(gcpPCRs)
		v.validators = []atls.Validator{
			gcp.NewValidator(gcpPCRs),
			gcp.NewNonCVMValidator(map[uint32][]byte{}), // TODO: Remove once we no longer use non CVMs.
		}
	case cloudprovider.Azure:
		azurePCRs := *config.Provider.Azure.PCRs
		if err := v.checkPCRs(azurePCRs); err != nil {
			return Validators{}, err
		}
		v.setPCRWarnings(azurePCRs)
		v.validators = []atls.Validator{
			azure.NewValidator(azurePCRs),
		}
	default:
		return Validators{}, errors.New("unsupported cloud provider")
	}
	return v, nil
}

// V returns validators as list of atls.Validator.
func (v *Validators) V() []atls.Validator {
	return v.validators
}

// Warnings returns warnings for the specifc PCR values that are not verified.
//
// PCR allocation inspired by https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12#Tab1
func (v *Validators) Warnings() string {
	return v.pcrWarnings
}

// WarningsIncludeInit returns warnings for the specifc PCR values that are not verified.
// Warnings regarding the initialization are included.
//
// PCR allocation inspired by https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12#Tab1
func (v *Validators) WarningsIncludeInit() string {
	return v.pcrWarnings + v.pcrWarningsInit
}

func (v *Validators) checkPCRs(pcrs map[uint32][]byte) error {
	for k, v := range pcrs {
		if len(v) != 32 {
			return fmt.Errorf("bad config: PCR[%d]: expected length: %d, but got: %d", k, 32, len(v))
		}
	}
	return nil
}

func (v *Validators) setPCRWarnings(pcrs map[uint32][]byte) {
	const warningStr = "Warning: not verifying the Constellation's %s measurements\n"
	sb := &strings.Builder{}

	if pcrs[0] == nil || pcrs[1] == nil {
		writeFmt(sb, warningStr, "BIOS")
	}

	if pcrs[2] == nil || pcrs[3] == nil {
		writeFmt(sb, warningStr, "OPROM")
	}

	if pcrs[4] == nil || pcrs[5] == nil {
		writeFmt(sb, warningStr, "MBR")
	}

	// GRUB measures kernel command line and initrd into pcrs 8 and 9
	if pcrs[8] == nil {
		writeFmt(sb, warningStr, "kernel command line")
	}
	if pcrs[9] == nil {
		writeFmt(sb, warningStr, "initrd")
	}
	v.pcrWarnings = sb.String()

	// Write init warnings separate.
	if pcrs[uint32(vtpm.PCRIndexOwnerID)] == nil || pcrs[uint32(vtpm.PCRIndexClusterID)] == nil {
		v.pcrWarningsInit = fmt.Sprintf(warningStr, "initialization status")
	}
}

func writeFmt(sb *strings.Builder, fmtStr string, args ...interface{}) {
	sb.WriteString(fmt.Sprintf(fmtStr, args...))
}
