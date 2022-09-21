/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package trustedlaunch

import (
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
)

// Issuer for Azure trusted launch TPM attestation.
type Issuer struct {
	oid.AzureTrustedLaunch
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
