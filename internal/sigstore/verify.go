/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"fmt"

	"github.com/sigstore/sigstore/pkg/cryptoutils"
	sigsig "github.com/sigstore/sigstore/pkg/signature"
)

// Verifier checks if the signature of content can be verified.
type Verifier interface {
	VerifySignature(content, signature []byte) error
}

// CosignVerifier wraps a public key that can be used for verifying signatures.
type CosignVerifier struct {
	publicKey crypto.PublicKey
}

// NewCosignVerifier unmarshalls and validates the given pem encoded public key and returns a new CosignVerifier.
func NewCosignVerifier(pem []byte) (Verifier, error) {
	pubkey, err := cryptoutils.UnmarshalPEMToPublicKey(pem)
	if err != nil {
		return CosignVerifier{}, fmt.Errorf("unable to parse public key: %w", err)
	}
	if err := cryptoutils.ValidatePubKey(pubkey); err != nil {
		return CosignVerifier{}, fmt.Errorf("unable to validate public key: %w", err)
	}

	return CosignVerifier{pubkey}, nil
}

// VerifySignature checks if the signature of content can be verified
// using publicKey.
// signature is expected to be base64 encoded.
// publicKey is expected to be PEM encoded.
func (c CosignVerifier) VerifySignature(content, signature []byte) error {
	// LoadVerifier would also error if no public key is set.
	// However, this error message should be easier to debug.
	if c.publicKey == nil {
		return fmt.Errorf("no public key set")
	}

	sigRaw := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(signature))

	verifier, err := sigsig.LoadVerifier(c.publicKey, crypto.SHA256)
	if err != nil {
		return fmt.Errorf("unable to load verifier: %w", err)
	}

	if err := verifier.VerifySignature(sigRaw, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("unable to verify signature: %w", err)
	}

	return nil
}

// IsBase64 checks if the given byte slice is base64 encoded.
func IsBase64(signature []byte) error {
	target := make([]byte, base64.StdEncoding.DecodedLen(len(signature)))
	_, err := base64.StdEncoding.Decode(target, signature)
	return err
}
