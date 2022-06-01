package azure

import (
	"crypto"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/google/go-tpm/tpm2"
)

// Validator for Azure confidential VM attestation.
type Validator struct {
	oid.Azure
	*vtpm.Validator
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(pcrs map[uint32][]byte) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			trustedKeyFromSNP,
			validateAzureCVM,
			vtpm.VerifyPKCS1v15,
		),
	}
}

// trustedKeyFromSNP establishes trust in the given public key.
// It does so by verifying the SNP attestation statement in instanceInfo.
//
// As long as we are using regular VMs on Azure this is a stub, only returning the given key.
func trustedKeyFromSNP(akPub, instanceInfo []byte) (crypto.PublicKey, error) {
	// TODO: convert this to SEV-SNP attestation verification
	pubArea, err := tpm2.DecodePublic(akPub)
	if err != nil {
		return nil, err
	}
	return pubArea.Key()
}

// validateAzureCVM validates Azure CVM capabilities.
//
// This might stay a stub, since SEV-SNP attestation is already verified in trustedKeyFromSNP().
func validateAzureCVM(attestation vtpm.AttestationDocument) error {
	// TODO: implement this for CVMs
	return nil
}
