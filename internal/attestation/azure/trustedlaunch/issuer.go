package trustedlaunch

import (
	"io"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
)

// Issuer for Azure trusted launch TPM attestation.
type Issuer struct {
	oid.Azure
	*vtpm.Issuer
}

// NewIssuer initializes a new Azure Issuer.
func NewIssuer() *Issuer {
	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			tpmclient.AttestationKeyRSA,
			getAttestation,
		),
	}
}

// getAttestation returns nil.
func getAttestation(tpm io.ReadWriteCloser) ([]byte, error) {
	return nil, nil
}
