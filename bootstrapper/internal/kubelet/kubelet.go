/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubelet

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"net"

	"github.com/edgelesssys/constellation/bootstrapper/internal/certificate"
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
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{constants.NodesGroup},
			CommonName:   constants.NodesUserPrefix + nodeName,
		},
		IPAddresses: ips,
	}
	return certificate.GetCertificateRequest(csrTemplate)
}
