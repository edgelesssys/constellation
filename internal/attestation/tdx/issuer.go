/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"encoding/json"

	"github.com/edgelesssys/constellation/v2/internal/oid"
)

type tdxIssuer interface {
	Close() error
	GenerateQuote(userData []byte) ([]byte, error)
}

// Issuer is the TDX attestation issuer.
type Issuer struct {
	oid.QEMUTDX

	open func() (tdxIssuer, error)
}

// NewIssuer initializes a new TDX Issuer.
func NewIssuer() *Issuer {
	return &Issuer{open: openTDXIssuer}
}

// Issue issues a TDX attestation document.
func (i *Issuer) Issue(userData []byte) ([]byte, error) {
	tdx, err := i.open()
	if err != nil {
		return nil, err
	}
	defer tdx.Close()

	quote, err := tdx.GenerateQuote(userData)
	if err != nil {
		return nil, err
	}

	return json.Marshal(tdxAttestationDocument{
		RawQuote: quote,
		UserData: userData,
	})
}

func openTDXIssuer() (tdxIssuer, error) {
	// return tdx.Open()
	return &tdxStub{}, nil
}

type tdxStub struct{}

func (t *tdxStub) Close() error { return nil }

func (t *tdxStub) GenerateQuote(userData []byte) ([]byte, error) { return nil, nil }
