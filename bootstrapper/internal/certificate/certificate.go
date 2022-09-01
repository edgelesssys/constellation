/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

// GetCertificateRequest returns a certificate request and matching private key.
func GetCertificateRequest(csrTemplate *x509.CertificateRequest) (certificateRequest []byte, privateKey []byte, err error) {
	privK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	keyBytes, err := x509.MarshalECPrivateKey(privK)
	if err != nil {
		return nil, nil, err
	}
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})
	certificateRequest, err = x509.CreateCertificateRequest(rand.Reader, csrTemplate, privK)
	if err != nil {
		return nil, nil, err
	}

	return certificateRequest, keyPem, nil
}
