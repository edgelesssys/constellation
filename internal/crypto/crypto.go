/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package crypto provides functions to for cryptography and random numbers.
package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"time"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/ssh"
)

const (
	// StateDiskKeyLength is key length in bytes for node state disk.
	StateDiskKeyLength = 32
	// DerivedKeyLengthDefault is the default length in bytes for KMS derived keys.
	DerivedKeyLengthDefault = 32
	// MasterSecretLengthDefault is the default length in bytes for CLI generated master secrets.
	MasterSecretLengthDefault = 32
	// MasterSecretLengthMin is the minimal length in bytes for user provided master secrets.
	MasterSecretLengthMin = 16
	// RNGLengthDefault is the number of bytes used for generating nonces.
	RNGLengthDefault = 32
	// DEKPrefix is the prefix used to prefix DEK IDs. Originally introduced as a requirement for the HKDF info parameter.
	DEKPrefix = "key-"
	// MeasurementSecretKeyID is name used for the measurementSecret DEK.
	MeasurementSecretKeyID = "measurementSecret"
)

// DeriveKey derives a key from a secret.
func DeriveKey(secret, salt, info []byte, length uint) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateCertificateSerialNumber generates a random serial number for an X.509 certificate.
func GenerateCertificateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

// GenerateRandomBytes reads length bytes from getrandom(2) if available, /dev/urandom otherwise.
func GenerateRandomBytes(length int) ([]byte, error) {
	nonce := make([]byte, length)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

// GenerateEmergencySSHCAKey creates a CA that is used to sign keys for emergency ssh access.
func GenerateEmergencySSHCAKey(seed []byte) (ssh.Signer, error) {
	_, priv, err := ed25519.GenerateKey(bytes.NewReader(seed))
	if err != nil {
		return nil, err
	}
	ca, err := ssh.NewSignerFromSigner(priv)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// GenerateSignedSSHHostKey creates a host key that is signed by the given CA.
func GenerateSignedSSHHostKey(principals []string, ca ssh.Signer) (*pem.Block, *ssh.Certificate, error) {
	hostKeyPub, hostKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}
	hostKeySSHPub, err := ssh.NewPublicKey(hostKeyPub)
	if err != nil {
		return nil, nil, err
	}
	pemHostKey, err := ssh.MarshalPrivateKey(hostKey, "")
	if err != nil {
		return nil, nil, err
	}

	certificate := ssh.Certificate{
		CertType:        ssh.HostCert,
		ValidPrincipals: principals,
		ValidAfter:      uint64(time.Now().Unix()),
		ValidBefore:     ssh.CertTimeInfinity,
		Reserved:        []byte{},
		Key:             hostKeySSHPub,
		KeyId:           principals[0],
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions:      map[string]string{},
		},
	}
	if err := certificate.SignCert(rand.Reader, ca); err != nil {
		return nil, nil, err
	}

	return pemHostKey, &certificate, nil
}

// PemToX509Cert takes a list of PEM-encoded certificates, parses the first one and returns it
// as an x.509 certificate.
func PemToX509Cert(raw []byte) (*x509.Certificate, error) {
	decoded, _ := pem.Decode(raw)
	if decoded == nil {
		return nil, fmt.Errorf("decoding pem: no PEM data found")
	}
	cert, err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing certificate: %w", err)
	}
	return cert, nil
}

// X509CertToPem takes an x.509 certificate and returns it as a PEM-encoded certificate.
func X509CertToPem(cert *x509.Certificate) ([]byte, error) {
	outWriter := &bytes.Buffer{}
	err := pem.Encode(outWriter, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return nil, fmt.Errorf("encode certificate: %w", err)
	}
	return outWriter.Bytes(), nil
}
