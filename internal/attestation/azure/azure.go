/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Azure attestation

Constellation supports multiple attestation technologies on Azure.

  - SEV - Secure Nested Paging (SEV-SNP)

    TPM attestation verified using an SEV-SNP attestation statement.

  - Trusted Launch

    Basic TPM attestation.
*/
package azure

import (
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
)

// NewIssuer returns an SNP issuer if it can successfully read the idkeydigest from the TPM.
// Otherwise returns a Trusted Launch issuer.
func NewIssuer(log vtpm.AttestationLogger) atls.Issuer {
	if _, err := snp.GetIDKeyDigest(vtpm.OpenVTPM); err == nil {
		return snp.NewIssuer(log)
	}
	return trustedlaunch.NewIssuer(log)
}
