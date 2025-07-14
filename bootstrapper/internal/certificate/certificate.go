/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package certificate provides functions to create a certificate request and matching private key.
package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"

	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

const (
	// CertificateFilename is the path to the kubelets certificate.
	CertificateFilename = "/run/state/kubelet/pki/kubelet-client-crt.pem"
	// KeyFilename is the path to the kubelets private key.
	KeyFilename = "/run/state/kubelet/pki/kubelet-client-key.pem"
)

// GetKubeletCertificateRequest returns a certificate request and matching private key for the kubelet.
func GetKubeletCertificateRequest(nodeName string, ips []net.IP) (certificateRequest []byte, privateKey []byte, err error) {
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{constants.NodesGroup},
			CommonName:   constants.NodesUserPrefix + nodeName,
		},
		IPAddresses: ips,
	}
	return GetCertificateRequest(csrTemplate)
}

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
