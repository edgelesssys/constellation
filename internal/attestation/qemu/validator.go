/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package qemu

import (
	"crypto"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/google/go-tpm/tpm2"
)

// Validator for QEMU VM attestation.
type Validator struct {
	oid.QEMU
	*vtpm.Validator
}

// NewValidator initializes a new QEMU validator with the provided PCR values.
func NewValidator(pcrs measurements.M, log attestation.Logger) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			unconditionalTrust,
			func(attestation vtpm.AttestationDocument) error { return nil },
			log,
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
