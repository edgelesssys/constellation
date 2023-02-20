/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"github.com/edgelesssys/constellation/v2/internal/oid"
)

// Issuer is the TDX attestation issuer.
type Issuer struct {
	oid.QEMUTDX
}

// NewIssuer initializes a new TDX Issuer.
func NewIssuer() *Issuer {
	return &Issuer{}
}
