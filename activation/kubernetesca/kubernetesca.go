package kubernetesca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/internal/file"
	"k8s.io/klog/v2"
)

const (
	caCertFilename = "/etc/kubernetes/pki/ca.crt"
	caKeyFilename  = "/etc/kubernetes/pki/ca.key"
)

// KubernetesCA handles signing of certificates using the Kubernetes root CA.
type KubernetesCA struct {
	file file.Handler
}

// New creates a new KubernetesCA.
func New(fileHandler file.Handler) *KubernetesCA {
	return &KubernetesCA{
		file: fileHandler,
	}
}

// GetCertificate creates a certificate for a node and signs it using the Kubernetes root CA.
func (c KubernetesCA) GetCertificate(nodeName string) (cert []byte, key []byte, err error) {
	klog.V(6).Info("CA: loading Kubernetes CA certificate")
	parentCertRaw, err := c.file.Read(caCertFilename)
	if err != nil {
		return nil, nil, err
	}
	parentCertPEM, _ := pem.Decode(parentCertRaw)
	parentCert, err := x509.ParseCertificate(parentCertPEM.Bytes)
	if err != nil {
		return nil, nil, err
	}

	klog.V(6).Info("CA: loading Kubernetes CA private key")
	parentKeyRaw, err := c.file.Read(caKeyFilename)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, fmt.Errorf("unsupported key type %q", parentCertPEM.Type)
	}
	if err != nil {
		return nil, nil, err
	}

	klog.V(6).Info("CA: creating kubelet private key")
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

	klog.V(6).Info("CA: creating kubelet certificate")
	serialNumber, err := util.GenerateCertificateSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	// Create the kubelet certificate
	// For a reference on the certificate fields, see: https://kubernetes.io/docs/setup/best-practices/certificates/
	certTmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    now.Add(-2 * time.Hour),
		NotAfter:     now.Add(24 * 365 * time.Hour),
		Subject: pkix.Name{
			Organization: []string{"system:nodes"},
			CommonName:   fmt.Sprintf("system:node:%s", nodeName),
		},
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
		IsCA:                  false,
		BasicConstraintsValid: true,
	}
	certRaw, err := x509.CreateCertificate(rand.Reader, certTmpl, parentCert, &privK.PublicKey, parentKey)
	if err != nil {
		return nil, nil, err
	}
	kubeletCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	return kubeletCert, kubeletKey, nil
}
