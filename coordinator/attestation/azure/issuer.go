package azure

import (
	"io"

	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
)

// Issuer for Azure TPM attestation.
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
			getSNPAttestation,
		),
	}
}

// getSNPAttestation loads and returns the SEV-SNP attestation statement.
//
// As long as we are using regular VMs on Azure this is a stub, returning nil.
func getSNPAttestation(tpm io.ReadWriteCloser) ([]byte, error) {
	// TODO: implement this for CVMs
	return nil, nil
}
