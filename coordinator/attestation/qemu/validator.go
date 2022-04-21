package qemu

import (
	"crypto"

	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/oid"
	"github.com/google/go-tpm/tpm2"
)

// Validator for QEMU VM attestation.
type Validator struct {
	oid.QEMU
	*vtpm.Validator
}

// NewValidator initializes a new qemu validator with the provided PCR values.
func NewValidator(pcrs map[uint32][]byte) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			unconditionalTrust,
			func(attestation vtpm.AttestationDocument) error { return nil },
			vtpm.VerifyPKCS1v15,
		),
	}
}

// unconditionalTrust returns the given public key as the trusted attestation key.
func unconditionalTrust(akPub, instanceInfo []byte) (crypto.PublicKey, error) {
	pubArea, err := tpm2.DecodePublic(akPub)
	if err != nil {
		return nil, err
	}
	return pubArea.Key()
}
