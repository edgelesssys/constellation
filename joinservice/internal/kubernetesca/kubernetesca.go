/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// kubernetesca implements a certificate authority that uses the Kubernetes root CA to sign certificates.
package kubernetesca

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

const (
	caCertFilename = "/etc/kubernetes/pki/ca.crt"
	caKeyFilename  = "/etc/kubernetes/pki/ca.key"
)

// KubernetesCA handles signing of certificates using the Kubernetes root CA.
type KubernetesCA struct {
	log  *slog.Logger
	file file.Handler
}

// New creates a new KubernetesCA.
func New(log *slog.Logger, fileHandler file.Handler) *KubernetesCA {
	return &KubernetesCA{
		log:  log,
		file: fileHandler,
	}
}

// GetNodeNameFromCSR extracts the node name from a CSR.
func (c KubernetesCA) GetNodeNameFromCSR(csr []byte) (string, error) {
	certRequest, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(certRequest.Subject.CommonName, kubeconstants.NodesUserPrefix) {
		return "", fmt.Errorf("certificate request must have common name prefix %q but is %q", kubeconstants.NodesUserPrefix, certRequest.Subject.CommonName)
	}

	return strings.TrimPrefix(certRequest.Subject.CommonName, kubeconstants.NodesUserPrefix), nil
}

// GetCertificate creates a certificate for a node and signs it using the Kubernetes root CA.
func (c KubernetesCA) GetCertificate(csr []byte) (cert []byte, err error) {
	c.log.Debug("Loading Kubernetes CA certificate")
	parentCertRaw, err := c.file.Read(caCertFilename)
	if err != nil {
		return nil, err
	}
	parentCert, err := crypto.PemToX509Cert(parentCertRaw)
	if err != nil {
		return nil, err
	}

	c.log.Debug("Loading Kubernetes CA private key")
	parentKeyRaw, err := c.file.Read(caKeyFilename)
	if err != nil {
		return nil, err
	}
	parentKeyPEM, _ := pem.Decode(parentKeyRaw)
	var parentKey any
	switch parentKeyPEM.Type {
	case "EC PRIVATE KEY":
		parentKey, err = x509.ParseECPrivateKey(parentKeyPEM.Bytes)
	case "RSA PRIVATE KEY":
		parentKey, err = x509.ParsePKCS1PrivateKey(parentKeyPEM.Bytes)
	case "PRIVATE KEY":
		parentKey, err = x509.ParsePKCS8PrivateKey(parentKeyPEM.Bytes)
	default:
		return nil, fmt.Errorf("unsupported key type %q", parentKeyPEM.Type)
	}
	if err != nil {
		return nil, err
	}

	certRequest, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return nil, err
	}
	if err := certRequest.CheckSignature(); err != nil {
		return nil, err
	}

	c.log.Info("Creating kubelet certificate")
	if len(certRequest.Subject.Organization) != 1 {
		return nil, errors.New("certificate request must have exactly one organization")
	}
	if certRequest.Subject.Organization[0] != kubeconstants.NodesGroup {
		return nil, fmt.Errorf("certificate request must have organization %q but has %q", kubeconstants.NodesGroup, certRequest.Subject.Organization[0])
	}
	if !strings.HasPrefix(certRequest.Subject.CommonName, kubeconstants.NodesUserPrefix) {
		return nil, fmt.Errorf("certificate request must have common name prefix %q but is %q", kubeconstants.NodesUserPrefix, certRequest.Subject.CommonName)
	}

	serialNumber, err := crypto.GenerateCertificateSerialNumber()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// Create the kubelet certificate
	// For a reference on the certificate fields, see: https://kubernetes.io/docs/setup/best-practices/certificates/
	certTmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    now.Add(-2 * time.Hour),
		NotAfter:     now.Add(24 * 365 * time.Hour),
		Subject:      certRequest.Subject,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IsCA:                  false,
		BasicConstraintsValid: true,
		IPAddresses:           certRequest.IPAddresses,
	}

	certRaw, err := x509.CreateCertificate(rand.Reader, certTmpl, parentCert, certRequest.PublicKey, parentKey)
	if err != nil {
		return nil, err
	}
	kubeletCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	return kubeletCert, nil
}
