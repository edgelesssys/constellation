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

// Product returns the SEV product info currently supported by Constellation's SNP attestation.
func Product() *spb.SevProduct {
	// sevProduct is the product info of the SEV platform as reported through CPUID[EAX=1].
	// It may become necessary in the future to differentiate among CSP vendors.
	return &spb.SevProduct{Name: spb.SevProduct_SEV_PRODUCT_MILAN, Stepping: 0} // Milan-B0
}

// InstanceInfo contains the necessary information to establish trust in a SNP CVM.
type InstanceInfo struct {
	// ReportSigner is the PEM-encoded certificate used to validate the attestation report's signature.
	ReportSigner []byte
	// CertChain is the PEM-encoded certificate chain for the attestation report (ASK+ARK).
	// Intermediate key that validates the ReportSigner and root key.
	CertChain []byte
	// AttestationReport is the attestation report from the vTPM (NVRAM) of the CVM.
	AttestationReport []byte
	Azure             *AzureInstanceInfo
}

// AzureInstanceInfo contains Azure specific information related to SNP attestation.
type AzureInstanceInfo struct {
	// RuntimeData is the Azure runtime data from the vTPM (NVRAM) of the CVM.
	RuntimeData []byte
	// MAAToken is the token of the MAA for the attestation report, used as a fallback
	// if the IDKeyDigest cannot be verified.
	MAAToken string
}

// addReportSigner parses the reportSigner certificate (VCEK/VLEK) from a and adds it to the attestation proto att.
// If reportSigner is empty and a VLEK is required, an error is returned.
// If reportSigner is empty and a VCEK is required, the VCEK is retrieved from AMD KDS.
func (a *InstanceInfo) addReportSigner(att *spb.Attestation, report *spb.Report, productName string, getter trust.HTTPSGetter, logger attestation.Logger) (abi.ReportSigner, error) {
	// If the VCEK certificate is present, parse it and format it.
	reportSigner, err := a.ParseReportSigner()
	if err != nil {
		logger.Warn("Error parsing report signer: %v", err)
	}

	signerInfo, err := abi.ParseSignerInfo(report.GetSignerInfo())
	if err != nil {
		return abi.NoneReportSigner, fmt.Errorf("parsing signer info: %w", err)
	}

	switch signerInfo.SigningKey {
	case abi.VlekReportSigner:
		if reportSigner == nil {
			return abi.NoneReportSigner, fmt.Errorf("VLEK certificate required but not present")
		}
		att.CertificateChain.VlekCert = reportSigner.Raw

	case abi.VcekReportSigner:
		var vcekData []byte

		// If no VCEK is present, fetch it from AMD.
		if reportSigner == nil {
			logger.Info("VCEK certificate not present, falling back to retrieving it from AMD KDS")
			vcekURL := kds.VCEKCertURL(productName, report.GetChipId(), kds.TCBVersion(report.GetReportedTcb()))
			vcekData, err = getter.Get(vcekURL)
			if err != nil {
				return abi.NoneReportSigner, fmt.Errorf("retrieving VCEK certificate from AMD KDS: %w", err)
			}
		} else {
			vcekData = reportSigner.Raw
		}

		att.CertificateChain.VcekCert = vcekData
	}

	return signerInfo.SigningKey, nil
}

// AttestationWithCerts returns a formatted version of the attestation report and its certificates from the instanceInfo.
// Certificates are retrieved in the following precedence:
// 1. ASK or ARK from issuer. On Azure: THIM. One AWS: not prefilled.
// 2. ASK or ARK from fallbackCerts.
// 3. ASK or ARK from AMD KDS.
func (a *InstanceInfo) AttestationWithCerts(getter trust.HTTPSGetter,
	fallbackCerts CertificateChain, logger attestation.Logger,
) (*spb.Attestation, error) {
	report, err := abi.ReportToProto(a.AttestationReport)
	if err != nil {
		return nil, fmt.Errorf("converting report to proto: %w", err)
	}

	productName := kds.ProductString(Product())

	att := &spb.Attestation{
		Report:           report,
		CertificateChain: &spb.CertificateChain{},
		Product:          Product(),
	}

	// Add VCEK/VLEK to attestation object.
	signingInfo, err := a.addReportSigner(att, report, productName, getter, logger)
	if err != nil {
		return nil, fmt.Errorf("adding report signer: %w", err)
	}

	// If the certificate chain from THIM is present, parse it and format it.
	ask, ark, err := a.ParseCertChain()
	if err != nil {
		logger.Warn("Error parsing certificate chain: %v", err)
	}
	if ask != nil {
		logger.Info("Using ASK certificate from Azure THIM")
		att.CertificateChain.AskCert = ask.Raw
	}
	if ark != nil {
		logger.Info("Using ARK certificate from Azure THIM")
		att.CertificateChain.ArkCert = ark.Raw
	}

	// If a cached ASK or an ARK from the Constellation config is present, use it.
	if att.CertificateChain.AskCert == nil && fallbackCerts.ask != nil {
		logger.Info("Using cached ASK certificate")
		att.CertificateChain.AskCert = fallbackCerts.ask.Raw
	}
	if att.CertificateChain.ArkCert == nil && fallbackCerts.ark != nil {
		logger.Info("Using ARK certificate from %s", constants.ConfigFilename)
		att.CertificateChain.ArkCert = fallbackCerts.ark.Raw
	}
	// Otherwise, retrieve it from AMD KDS.
	if att.CertificateChain.AskCert == nil || att.CertificateChain.ArkCert == nil {
		logger.Info(
			"Certificate chain not fully present (ARK present: %t, ASK present: %t), falling back to retrieving it from AMD KDS",
			(att.CertificateChain.ArkCert != nil),
			(att.CertificateChain.AskCert != nil),
		)
		kdsCertChain, err := trust.GetProductChain(productName, signingInfo, getter)
		if err != nil {
			return nil, fmt.Errorf("retrieving certificate chain from AMD KDS: %w", err)
		}
		if att.CertificateChain.AskCert == nil && kdsCertChain.Ask != nil {
			logger.Info("Using ASK certificate from AMD KDS")
			att.CertificateChain.AskCert = kdsCertChain.Ask.Raw
		}
		if att.CertificateChain.ArkCert == nil && kdsCertChain.Ask != nil {
			logger.Info("Using ARK certificate from AMD KDS")
			att.CertificateChain.ArkCert = kdsCertChain.Ark.Raw
		}
	}

	return att, nil
}

// CertificateChain stores an AMD signing key (ASK) and AMD root key (ARK) certificate.
type CertificateChain struct {
	ask *x509.Certificate
	ark *x509.Certificate
}

// NewCertificateChain returns a new CertificateChain with the given ASK and ARK certificates.
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
		case "SEV-Milan", "SEV-VLEK-Milan":
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

// ParseReportSigner parses the VCEK/VLEK certificate from the instanceInfo into an x509-formatted certificate.
// If no certificate is present, nil is returned.
func (a *InstanceInfo) ParseReportSigner() (*x509.Certificate, error) {
	newlinesTrimmed := bytes.TrimSpace(a.ReportSigner)
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

	reportSigner, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing VCEK certificate: %w", err)
	}

	return reportSigner, nil
}
