/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package sigstore

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/secure-systems-lab/go-securesystemslib/encrypted"
	"github.com/sigstore/sigstore/pkg/signature"
)

const (
	cosignPrivateKeyPemType   = "ENCRYPTED COSIGN PRIVATE KEY"
	sigstorePrivateKeyPemType = "ENCRYPTED SIGSTORE PRIVATE KEY"
)

// Signer is used to sign the version file. Used for unit testing.
type Signer interface {
	Sign(content []byte) (res []byte, err error)
}

// NewSigner returns a new Signer.
func NewSigner(cosignPwd, privKey []byte) Signer {
	return signer{cosignPwd: cosignPwd, privKey: privKey}
}

type signer struct {
	cosignPwd []byte // used to decrypt the cosign private key
	privKey   []byte // used to sign
}

func (s signer) Sign(content []byte) (signature []byte, err error) {
	signature, err = SignContent(s.cosignPwd, s.privKey, content)
	if err != nil {
		return signature, fmt.Errorf("sign version file: %w", err)
	}
	return
}

// SignContent signs the content with the cosign encrypted private key and corresponding cosign password.
func SignContent(password, encryptedPrivateKey, content []byte) ([]byte, error) {
	sv, err := loadPrivateKey(encryptedPrivateKey, password)
	if err != nil {
		return nil, fmt.Errorf("loading private key: %w", err)
	}
	sig, err := sv.SignMessage(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("signing message: %w", err)
	}
	return []byte(base64.StdEncoding.EncodeToString(sig)), nil
}

func loadPrivateKey(key []byte, pass []byte) (signature.SignerVerifier, error) {
	// Decrypt first
	p, _ := pem.Decode(key)
	if p == nil {
		return nil, errors.New("invalid pem block")
	}
	if p.Type != cosignPrivateKeyPemType && p.Type != sigstorePrivateKeyPemType {
		return nil, fmt.Errorf("unsupported pem type: %s", p.Type)
	}

	x509Encoded, err := encrypted.Decrypt(p.Bytes, pass)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	pk, err := x509.ParsePKCS8PrivateKey(x509Encoded)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}
	switch pk := pk.(type) {
	case *rsa.PrivateKey:
		return signature.LoadRSAPKCS1v15SignerVerifier(pk, crypto.SHA256)
	case *ecdsa.PrivateKey:
		return signature.LoadECDSASignerVerifier(pk, crypto.SHA256)
	case ed25519.PrivateKey:
		return signature.LoadED25519SignerVerifier(pk)
	default:
		return nil, errors.New("unsupported key type")
	}
}
