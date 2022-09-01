package azure

import (
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
)

// IsCVM checks if it can open a vTPM and read from the NV index expected to be available on CVMs without error.
func IsCVM(open vtpm.TPMOpenFunc) bool {
	_, err := snp.GetIdKeyDigest(open)
	return err == nil
}

func NewIssuer() atls.Issuer {
	if IsCVM(vtpm.OpenVTPM) {
		return snp.NewIssuer()
	} else {
		return trustedlaunch.NewIssuer()
	}
}
