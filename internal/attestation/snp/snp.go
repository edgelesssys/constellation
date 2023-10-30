/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package SNP provides types shared by SNP-based attestation implementations.
// It ensures all issuers provide the same types to the verify command.
package snp

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify/trust"
)

// InstanceInfo contains the necessary information to establish trust in
// an Azure CVM.
type InstanceInfo struct {
	// VCEK is the PEM-encoded VCEK certificate for the attestation report.
	VCEK []byte
	// CertChain is the PEM-encoded certificate chain for the attestation report.
	CertChain []byte
	// AttestationReport is the attestation report from the vTPM (NVRAM) of the CVM.
	AttestationReport []byte
	// RuntimeData is the Azure runtime data from the vTPM (NVRAM) of the CVM.
	RuntimeData []byte
	// MAAToken is the token of the MAA for the attestation report, used as a fallback
	// if the IDKeyDigest cannot be verified.
	MAAToken string
}

// AttestationWithCerts returns a formatted version of the attestation report and its certificates from the instanceInfo.
// Certificates are retrieved in the following precedence:
// 1. ASK or ARK from issuer. On Azure: THIM. One AWS: not prefilled.
// 2. ASK or ARK from fallbackCerts.
// 3. ASK or ARK from AMD KDS.
func (a *InstanceInfo) AttestationWithCerts(logger attestation.Logger, getter trust.HTTPSGetter,
	fallbackCerts CertificateChain,
) (*spb.Attestation, error) {
	report, err := abi.ReportToProto(a.AttestationReport)
	if err != nil {
		return nil, fmt.Errorf("converting report to proto: %w", err)
	}

	// Product info as reported through CPUID[EAX=1]
	sevProduct := &spb.SevProduct{Name: spb.SevProduct_SEV_PRODUCT_MILAN, Stepping: 0} // Milan-B0
	productName := kds.ProductString(sevProduct)

	att := &spb.Attestation{
		Report:           report,
		CertificateChain: &spb.CertificateChain{},
		Product:          sevProduct,
	}

	// If the VCEK certificate is present, parse it and format it.
	vcek, err := a.ParseVCEK()
	if err != nil {
		logger.Warnf("Error parsing VCEK: %v", err)
	}
	if vcek != nil {
		att.CertificateChain.VcekCert = vcek.Raw
	} else {
		// Otherwise, retrieve it from AMD KDS.
		logger.Infof("VCEK certificate not present, falling back to retrieving it from AMD KDS")
		vcekURL := kds.VCEKCertURL(productName, report.GetChipId(), kds.TCBVersion(report.GetReportedTcb()))
		vcek, err := getter.Get(vcekURL)
		if err != nil {
			return nil, fmt.Errorf("retrieving VCEK certificate from AMD KDS: %w", err)
		}
		att.CertificateChain.VcekCert = vcek
	}

	// If the certificate chain from THIM is present, parse it and format it.
	ask, ark, err := a.ParseCertChain()
	if err != nil {
		logger.Warnf("Error parsing certificate chain: %v", err)
	}
	if ask != nil {
		logger.Infof("Using ASK certificate from Azure THIM")
		att.CertificateChain.AskCert = ask.Raw
	}
	if ark != nil {
		logger.Infof("Using ARK certificate from Azure THIM")
		att.CertificateChain.ArkCert = ark.Raw
	}

	// If a cached ASK or an ARK from the Constellation config is present, use it.
	if att.CertificateChain.AskCert == nil && fallbackCerts.ask != nil {
		logger.Infof("Using cached ASK certificate")
		att.CertificateChain.AskCert = fallbackCerts.ask.Raw
	}
	if att.CertificateChain.ArkCert == nil && fallbackCerts.ark != nil {
		logger.Infof("Using ARK certificate from %s", constants.ConfigFilename)
		att.CertificateChain.ArkCert = fallbackCerts.ark.Raw
	}
	// Otherwise, retrieve it from AMD KDS.
	if att.CertificateChain.AskCert == nil || att.CertificateChain.ArkCert == nil {
		logger.Infof(
			"Certificate chain not fully present (ARK present: %t, ASK present: %t), falling back to retrieving it from AMD KDS",
			(att.CertificateChain.ArkCert != nil),
			(att.CertificateChain.AskCert != nil),
		)
		kdsCertChain, err := trust.GetProductChain(productName, abi.VcekReportSigner, getter)
		if err != nil {
			return nil, fmt.Errorf("retrieving certificate chain from AMD KDS: %w", err)
		}
		if att.CertificateChain.AskCert == nil && kdsCertChain.Ask != nil {
			logger.Infof("Using ASK certificate from AMD KDS")
			att.CertificateChain.AskCert = kdsCertChain.Ask.Raw
		}
		if att.CertificateChain.ArkCert == nil && kdsCertChain.Ask != nil {
			logger.Infof("Using ARK certificate from AMD KDS")
			att.CertificateChain.ArkCert = kdsCertChain.Ark.Raw
		}
	}

	return att, nil
}

type CertificateChain struct {
	ask *x509.Certificate
	ark *x509.Certificate
}

func NewCertificateChain(ask, ark *x509.Certificate) CertificateChain {
	return CertificateChain{
		ask: ask,
		ark: ark,
	}
}

// ParseCertChain parses the certificate chain from the instanceInfo into x509-formatted ASK and ARK certificates.
// If less than 2 certificates are present, only the present certificate is returned.
// If more than 2 certificates are present, an error is returned.
func (a *InstanceInfo) ParseCertChain() (ask, ark *x509.Certificate, retErr error) {
	rest := bytes.TrimSpace(a.CertChain)

	i := 1
	var block *pem.Block
	for block, rest = pem.Decode(rest); block != nil; block, rest = pem.Decode(rest) {
		if i > 2 {
			retErr = fmt.Errorf("parse certificate %d: more than 2 certificates in chain", i)
			return
		}

		if block.Type != "CERTIFICATE" {
			retErr = fmt.Errorf("parse certificate %d: expected PEM block type 'CERTIFICATE', got '%s'", i, block.Type)
			return
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			retErr = fmt.Errorf("parse certificate %d: %w", i, err)
			return
		}

		// https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/57230.pdf
		// Table 6 and 7
		switch cert.Subject.CommonName {
		case "SEV-Milan":
			ask = cert
		case "ARK-Milan":
			ark = cert
		default:
			retErr = fmt.Errorf("parse certificate %d: unexpected subject CN %s", i, cert.Subject.CommonName)
			return
		}

		i++
	}

	switch {
	case i == 1:
		retErr = fmt.Errorf("no PEM blocks found")
	case len(rest) != 0:
		retErr = fmt.Errorf("remaining PEM block is not a valid certificate: %s", rest)
	}

	return
}

// ParseVCEK parses the VCEK certificate from the instanceInfo into an x509-formatted certificate.
// If the VCEK certificate is not present, nil is returned.
func (a *InstanceInfo) ParseVCEK() (*x509.Certificate, error) {
	newlinesTrimmed := bytes.TrimSpace(a.VCEK)
	if len(newlinesTrimmed) == 0 {
		// VCEK is not present.
		return nil, nil
	}

	block, rest := pem.Decode(newlinesTrimmed)
	if block == nil {
		return nil, fmt.Errorf("no PEM blocks found")
	}
	if len(rest) != 0 {
		return nil, fmt.Errorf("received more data than expected")
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("expected PEM block type 'CERTIFICATE', got '%s'", block.Type)
	}

	vcek, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing VCEK certificate: %w", err)
	}

	return vcek, nil
}
