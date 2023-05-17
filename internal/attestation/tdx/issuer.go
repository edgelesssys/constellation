/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/go-tdx-qpl/tdx"
)

// Issuer is the TDX attestation issuer.
type Issuer struct {
	variant.QEMUTDX

	open OpenFunc
	log  attestation.Logger
}

// NewIssuer initializes a new TDX Issuer.
func NewIssuer(log attestation.Logger) *Issuer {
	if log == nil {
		log = attestation.NOPLogger{}
	}
	return &Issuer{
		open: Open,
		log:  log,
	}
}

// Issue issues a TDX attestation document.
func (i *Issuer) Issue(_ context.Context, userData []byte, nonce []byte) (attDoc []byte, err error) {
	i.log.Infof("Issuing attestation statement")
	defer func() {
		if err != nil {
			i.log.Warnf("Failed to issue attestation document: %s", err)
		}
	}()

	handle, err := i.open()
	if err != nil {
		return nil, fmt.Errorf("opening TDX device: %w", err)
	}
	defer handle.Close()

	quote, err := tdx.GenerateQuote(handle, attestation.MakeExtraData(userData, nonce))
	if err != nil {
		return nil, fmt.Errorf("generating quote: %w", err)
	}

	rawAttDoc, err := json.Marshal(tdxAttestationDocument{
		RawQuote: quote,
		UserData: userData,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling attestation document: %w", err)
	}

	return rawAttDoc, nil
}
