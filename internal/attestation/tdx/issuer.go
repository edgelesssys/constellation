/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"encoding/json"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/edgelesssys/go-tdx-qpl/tdx"
)

// Issuer is the TDX attestation issuer.
type Issuer struct {
	oid.QEMUTDX

	open OpenFunc
}

// NewIssuer initializes a new TDX Issuer.
func NewIssuer(open OpenFunc) *Issuer {
	return &Issuer{open: open}
}

// Issue issues a TDX attestation document.
func (i *Issuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	handle, err := i.open()
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	quote, err := tdx.GenerateQuote(handle, attestation.MakeExtraData(userData, nonce))
	if err != nil {
		return nil, err
	}

	return json.Marshal(tdxAttestationDocument{
		RawQuote: quote,
		UserData: userData,
	})
}
