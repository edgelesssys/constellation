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

// VerifySignature checks if the signature of content can be verified
// using publicKey.
// signature is expected to be base64 encoded.
// publicKey is expected to be PEM encoded.
func VerifySignature(content, signature, publicKey []byte) error {
	sigRaw := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(signature))

	pubKeyRaw, err := cryptoutils.UnmarshalPEMToPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("unable to parse public key: %w", err)
	}
	if err := cryptoutils.ValidatePubKey(pubKeyRaw); err != nil {
		return fmt.Errorf("unable to validate public key: %w", err)
	}

	verifier, err := sigsig.LoadVerifier(pubKeyRaw, crypto.SHA256)
	if err != nil {
		return fmt.Errorf("unable to load verifier: %w", err)
	}

	if err := verifier.VerifySignature(sigRaw, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("unable to verify signature: %w", err)
	}

	return nil
}
