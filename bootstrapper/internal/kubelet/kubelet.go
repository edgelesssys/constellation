/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubelet

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

// GetCertificateRequest returns a certificate request and macthing private key for the kubelet.
func GetCertificateRequest(nodeName string, ips []net.IP) (certificateRequest []byte, privateKey []byte, err error) {
	privK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	keyBytes, err := x509.MarshalECPrivateKey(privK)
	if err != nil {
		return nil, nil, err
	}
	kubeletKey := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{constants.NodesGroup},
			CommonName:   constants.NodesUserPrefix + nodeName,
		},
		IPAddresses: ips,
	}
	certificateRequest, err = x509.CreateCertificateRequest(rand.Reader, csrTemplate, privK)
	if err != nil {
		return nil, nil, err
	}

	return certificateRequest, kubeletKey, nil
}
