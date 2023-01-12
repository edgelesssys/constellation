/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"

	"github.com/google/tink/go/kwp/subtle"
)

// WrapAES performs AES Key Wrap with Padding as specified in RFC 5649: https://datatracker.ietf.org/doc/html/rfc5649
//
// Key sizes are limited to 16 and 32 Bytes.
func WrapAES(key []byte, wrapKeyAES []byte) ([]byte, error) {
	wrapper, err := subtle.NewKWP(wrapKeyAES)
	if err != nil {
		return nil, err
	}

	return wrapper.Wrap(key)
}

// UnwrapAES decrypts data wrapped with AES Key Wrap with Padding as specified in RFC 5649: https://datatracker.ietf.org/doc/html/rfc5649
//
// Key sizes are limited to 16 and 32 Bytes.
func UnwrapAES(encryptedKey []byte, wrapKeyAES []byte) ([]byte, error) {
	wrapper, err := subtle.NewKWP(wrapKeyAES)
	if err != nil {
		return nil, err
	}
	return wrapper.Unwrap(encryptedKey)
}

// ParsePEMtoPublicKeyRSA parses a public RSA key from bytes to *rsa.PublicKey.
func ParsePEMtoPublicKeyRSA(pkPEM []byte) (*rsa.PublicKey, error) {
	pkDER, _ := pem.Decode(pkPEM)
	if pkDER == nil {
		return nil, fmt.Errorf("did not find any PEM data")
	}
	if pkDER.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM format: want [PUBLIC KEY], got [%s]", pkDER.Type)
	}
	return ParseDERtoPublicKeyRSA(pkDER.Bytes)
}

// ParseDERtoPublicKeyRSA parses a PKIX, ASN.1 DER RSA public key from []byte to *rsa.PublicKey.
func ParseDERtoPublicKeyRSA(pkDER []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(pkDER)
	if err != nil {
		return nil, err
	}
	switch t := key.(type) {
	case *rsa.PublicKey:
		return t, nil
	default:
		return nil, fmt.Errorf("invalid key type: want [*rsa.PublicKey], got [%v]", reflect.TypeOf(t))
	}
}

// GetRandomKey reads length bytes from getrandom(2) if available, /dev/urandom otherwise.
func GetRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
