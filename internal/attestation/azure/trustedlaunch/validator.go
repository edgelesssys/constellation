/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package trustedlaunch

import (
	"crypto"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/google/go-tpm/tpm2"
)

// Validator for Azure trusted launch VM attestation.
type Validator struct {
	oid.AzureTrustedLaunch
	*vtpm.Validator
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(pcrs map[uint32][]byte, enforcedPCRs []uint32, log vtpm.WarnLogger) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			enforcedPCRs,
			trustedKey,
			validateVM,
			vtpm.VerifyPKCS1v15,
			log,
		),
	}
}

// trustedKey returns the key encoded in the given TPMT_PUBLIC message.
func trustedKey(akPub, instanceInfo []byte) (crypto.PublicKey, error) {
	pubArea, err := tpm2.DecodePublic(akPub)
	if err != nil {
		return nil, err
	}
	return pubArea.Key()
}

// validateVM returns nil.
func validateVM(attestation vtpm.AttestationDocument) error {
	return nil
}
