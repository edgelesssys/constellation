package azure

import (
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
)

func NewValidator(m map[uint32][]byte, e []uint32, idkeydigest []byte, enforceIdKeyDigest bool, log vtpm.WarnLogger) atls.Validator {
	if snp.IsCVM(vtpm.OpenVTPM) {
		return snp.NewValidator(m, e, idkeydigest, enforceIdKeyDigest, log)
	} else {
		return trustedlaunch.NewValidator(m, e, log)
	}
}
