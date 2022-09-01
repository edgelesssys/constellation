package azure

import (
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
)

func NewIssuer() atls.Issuer {
	if snp.IsCVM(vtpm.OpenVTPM) {
		return snp.NewIssuer()
	} else {
		return trustedlaunch.NewIssuer()
	}
}
