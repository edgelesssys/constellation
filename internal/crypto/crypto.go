// Package crypto provides functions to for cryptography and random numbers.
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/hkdf"
)

const (
	StateDiskKeyLength = 32
	// DerivedKeyLengthDefault is the default length in bytes for KMS derived keys.
	DerivedKeyLengthDefault = 32
	// MasterSecretLengthDefault is the default length in bytes for CLI generated master secrets.
	MasterSecretLengthDefault = 32
	// MasterSecretLengthMin is the minimal length in bytes for user provided master secrets.
	MasterSecretLengthMin = 16
	// RNGLengthDefault is the number of bytes used for generating nonces.
	RNGLengthDefault = 32
	// HKDFInfoPrefix is the prefix used for the info parameter in HKDF.
	HKDFInfoPrefix = "key-"
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

// PemToX509Cert takes a list of PEM encoded certificates, parses the first one and returns it.
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
