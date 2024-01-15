/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package verify provides the types for the verify report in JSON format.

The package provides an interface for constellation verify and
the attestationconfigapi upload tool through JSON serialization.
It exposes a CSP-agnostic interface for printing Reports that may include CSP-specific information.
*/
package verify

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	"github.com/google/go-sev-guest/verify/trust"
)

const (
	vcekCert         = "VCEK certificate"
	vlekCert         = "VLEK certificate"
	certificateChain = "certificate chain"
)

// Report contains the entire data reported by constellation verify.
type Report struct {
	SNPReport            SNPReport     `json:"snp_report"`
	ReportSigner         []Certificate `json:"report_signer"`
	CertChain            []Certificate `json:"cert_chain"`
	*AzureReportAddition `json:"azure,omitempty"`
	*AWSReportAddition   `json:"aws,omitempty"`
}

// AzureReportAddition contains attestation report data specific to Azure.
type AzureReportAddition struct {
	MAAToken MaaTokenClaims `json:"maa_token"`
}

// AWSReportAddition contains attestation report data specific to AWS.
type AWSReportAddition struct{}

// NewReport transforms a snp.InstanceInfo object into a Report.
func NewReport(ctx context.Context, instanceInfo snp.InstanceInfo, attestationCfg config.AttestationCfg, log debugLog) (Report, error) {
	snpReport, err := newSNPReport(instanceInfo.AttestationReport)
	if err != nil {
		return Report{}, fmt.Errorf("parsing SNP report: %w", err)
	}

	var certTypeName string
	switch snpReport.SignerInfo.SigningKey {
	case abi.VlekReportSigner.String():
		certTypeName = vlekCert
	case abi.VcekReportSigner.String():
		certTypeName = vcekCert
	default:
		return Report{}, errors.New("unknown report signer")
	}

	reportSigner, err := newCertificates(certTypeName, instanceInfo.ReportSigner, log)
	if err != nil {
		return Report{}, fmt.Errorf("parsing %s: %w", certTypeName, err)
	}

	// check if issuer included certChain before parsing. If not included, manually collect from the cluster.
	rawCerts := instanceInfo.CertChain
	if certTypeName == vlekCert {
		rawCerts, err = getCertChain(attestationCfg)
		if err != nil {
			return Report{}, fmt.Errorf("getting certificate chain cache: %w", err)
		}
	}

	certChain, err := newCertificates(certificateChain, rawCerts, log)
	if err != nil {
		return Report{}, fmt.Errorf("parsing certificate chain: %w", err)
	}

	var azure *AzureReportAddition
	var aws *AWSReportAddition
	if instanceInfo.Azure != nil {
		cfg, ok := attestationCfg.(*config.AzureSEVSNP)
		if !ok {
			return Report{}, fmt.Errorf("expected config type *config.AzureSEVSNP, got %T", attestationCfg)
		}
		maaToken, err := newMAAToken(ctx, instanceInfo.Azure.MAAToken, cfg.FirmwareSignerConfig.MAAURL)
		if err != nil {
			return Report{}, fmt.Errorf("parsing MAA token: %w", err)
		}
		azure = &AzureReportAddition{
			MAAToken: maaToken,
		}
	}

	return Report{
		SNPReport:           snpReport,
		ReportSigner:        reportSigner,
		CertChain:           certChain,
		AzureReportAddition: azure,
		AWSReportAddition:   aws,
	}, nil
}

// inverse of newCertificates.
// ideally, duplicate encoding/decoding would be removed.
// AWS specific.
func getCertChain(cfg config.AttestationCfg) ([]byte, error) {
	awsCfg, ok := cfg.(*config.AWSSEVSNP)
	if !ok {
		return nil, fmt.Errorf("expected config type *config.AWSSEVSNP, got %T", cfg)
	}

	if awsCfg.AMDRootKey.Equal(config.Certificate{}) {
		return nil, errors.New("no AMD root key configured")
	}

	if awsCfg.AMDSigningKey.Equal(config.Certificate{}) {
		certs, err := trust.GetProductChain(kds.ProductString(snp.Product()), abi.VlekReportSigner, trust.DefaultHTTPSGetter())
		if err != nil {
			return nil, fmt.Errorf("getting product certificate chain: %w", err)
		}
		// we want an ASVK, but GetProductChain currently does not use the ASVK field.
		if certs.Ask == nil {
			return nil, errors.New("no ASVK certificate available")
		}
		awsCfg.AMDSigningKey = config.Certificate(*certs.Ask)
	}

	// ARK
	certChain := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: awsCfg.AMDRootKey.Raw,
	})

	// append ASK
	certChain = append(certChain, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: awsCfg.AMDSigningKey.Raw,
	})...)

	return certChain, nil
}

// FormatString builds a string representation of a report that is inteded for console output.
func (r *Report) FormatString(b *strings.Builder) (string, error) {
	if len(r.ReportSigner) != 1 {
		return "", fmt.Errorf("expected exactly one report signing certificate, found %d", len(r.ReportSigner))
	}

	if err := formatCertificates(b, r.ReportSigner); err != nil {
		return "", fmt.Errorf("building report signing certificate string: %w", err)
	}

	if err := formatCertificates(b, r.CertChain); err != nil {
		return "", fmt.Errorf("building certificate chain string: %w", err)
	}

	r.SNPReport.formatString(b)
	if r.AzureReportAddition != nil {
		if err := r.AzureReportAddition.MAAToken.formatString(b); err != nil {
			return "", fmt.Errorf("error building MAAToken string : %w", err)
		}
	}

	return b.String(), nil
}

func formatCertificates(b *strings.Builder, certs []Certificate) error {
	for i, cert := range certs {
		if i == 0 {
			b.WriteString(fmt.Sprintf("\tRaw %s:\n", cert.CertTypeName))
		}
		newlinesTrimmed := strings.TrimSpace(cert.CertificatePEM)
		formattedCert := strings.ReplaceAll(newlinesTrimmed, "\n", "\n\t\t") + "\n"
		b.WriteString(fmt.Sprintf("\t\t%s", formattedCert))
	}
	for i, cert := range certs {
		// Use 1-based indexing for user output.
		if err := cert.formatString(b, i+1); err != nil {
			return fmt.Errorf("error printing certificate chain: %w", err)
		}
	}

	return nil
}

// Certificate contains the certificate data and additional information.
type Certificate struct {
	x509.Certificate `json:"-"`
	CertificatePEM   string     `json:"certificate"`
	CertTypeName     string     `json:"cert_type_name"`
	StructVersion    uint8      `json:"struct_version"`
	ProductName      string     `json:"product_name"`
	HardwareID       []byte     `json:"hardware_id,omitempty"`
	CspID            string     `json:"csp_id,omitempty"`
	TCBVersion       TCBVersion `json:"tcb_version"`
}

// newCertificates parses a list of PEM encoded certificate and returns a slice of Certificate objects.
func newCertificates(certTypeName string, cert []byte, log debugLog) (certs []Certificate, err error) {
	newlinesTrimmed := strings.TrimSpace(string(cert))

	log.Debug(fmt.Sprintf("Decoding PEM certificate: %s", certTypeName))
	i := 1
	var rest []byte
	var block *pem.Block
	for block, rest = pem.Decode([]byte(newlinesTrimmed)); block != nil; block, rest = pem.Decode(rest) {
		log.Debug(fmt.Sprintf("Parsing PEM block: %d", i))
		if block.Type != "CERTIFICATE" {
			return certs, fmt.Errorf("parse %s: expected PEM block type 'CERTIFICATE', got '%s'", certTypeName, block.Type)
		}

		certX509, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, fmt.Errorf("parse %s: %w", certTypeName, err)
		}

		var ext *kds.Extensions
		switch certTypeName {
		case vcekCert:
			ext, err = kds.VcekCertificateExtensions(certX509)
			if err != nil {
				return certs, fmt.Errorf("parsing %s extensions: %w", certTypeName, err)
			}
		case vlekCert:
			ext, err = kds.VlekCertificateExtensions(certX509)
			if err != nil {
				return certs, fmt.Errorf("parsing %s extensions: %w", certTypeName, err)
			}
		}
		certPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certX509.Raw,
		})
		cert := Certificate{
			Certificate:    *certX509,
			CertificatePEM: string(certPEM),
			CertTypeName:   certTypeName,
		}

		if ext != nil {
			cert.StructVersion = ext.StructVersion
			cert.ProductName = ext.ProductName
			cert.TCBVersion = newTCBVersion(ext.TCBVersion)
			if ext.HWID != nil {
				cert.HardwareID = ext.HWID
			} else {
				cert.CspID = ext.CspID
			}
		}

		certs = append(certs, cert)

		i++
	}
	if i == 1 {
		return certs, fmt.Errorf("parse %s: no PEM blocks found", certTypeName)
	}
	if len(rest) != 0 {
		return certs, fmt.Errorf("parse %s: remaining PEM block is not a valid certificate: %s", certTypeName, rest)
	}
	return certs, nil
}

// formatString builds a string representation of a certificate that is inteded for console output.
func (c *Certificate) formatString(b *strings.Builder, idx int) error {
	writeIndentfln(b, 1, "%s (%d):", c.CertTypeName, idx)
	writeIndentfln(b, 2, "Serial Number: %s", c.Certificate.SerialNumber)
	writeIndentfln(b, 2, "Subject: %s", c.Certificate.Subject)
	writeIndentfln(b, 2, "Issuer: %s", c.Certificate.Issuer)
	writeIndentfln(b, 2, "Not Before: %s", c.Certificate.NotBefore)
	writeIndentfln(b, 2, "Not After: %s", c.Certificate.NotAfter)
	writeIndentfln(b, 2, "Signature Algorithm: %s", c.Certificate.SignatureAlgorithm)
	writeIndentfln(b, 2, "Public Key Algorithm: %s", c.Certificate.PublicKeyAlgorithm)

	if c.CertTypeName == vcekCert {
		// Extensions documented in Table 8 and Table 9 of
		// https://www.amd.com/system/files/TechDocs/57230.pdf
		vcekExts, err := kds.VcekCertificateExtensions(&c.Certificate)
		if err != nil {
			return fmt.Errorf("parsing %s extensions: %w", c.CertTypeName, err)
		}

		writeIndentfln(b, 2, "Struct version: %d", vcekExts.StructVersion)
		writeIndentfln(b, 2, "Product name: %s", vcekExts.ProductName)
		tcb := kds.DecomposeTCBVersion(vcekExts.TCBVersion)
		writeIndentfln(b, 2, "Secure Processor bootloader SVN: %d", tcb.BlSpl)
		writeIndentfln(b, 2, "Secure Processor operating system SVN: %d", tcb.TeeSpl)
		writeIndentfln(b, 2, "SVN 4 (reserved): %d", tcb.Spl4)
		writeIndentfln(b, 2, "SVN 5 (reserved): %d", tcb.Spl5)
		writeIndentfln(b, 2, "SVN 6 (reserved): %d", tcb.Spl6)
		writeIndentfln(b, 2, "SVN 7 (reserved): %d", tcb.Spl7)
		writeIndentfln(b, 2, "SEV-SNP firmware SVN: %d", tcb.SnpSpl)
		writeIndentfln(b, 2, "Microcode SVN: %d", tcb.UcodeSpl)
		writeIndentfln(b, 2, "Hardware ID: %x", vcekExts.HWID)
	}

	return nil
}

// TCBVersion contains the TCB version data.
type TCBVersion struct {
	Bootloader uint8 `json:"bootloader"`
	TEE        uint8 `json:"tee"`
	SNP        uint8 `json:"snp"`
	Microcode  uint8 `json:"microcode"`
	Spl4       uint8 `json:"spl4"`
	Spl5       uint8 `json:"spl5"`
	Spl6       uint8 `json:"spl6"`
	Spl7       uint8 `json:"spl7"`
}

// formatString builds a string representation of a TCB version that is inteded for console output.
func (t *TCBVersion) formatString(b *strings.Builder) {
	writeIndentfln(b, 3, "Secure Processor bootloader SVN: %d", t.Bootloader)
	writeIndentfln(b, 3, "Secure Processor operating system SVN: %d", t.TEE)
	writeIndentfln(b, 3, "SVN 4 (reserved): %d", t.Spl4)
	writeIndentfln(b, 3, "SVN 5 (reserved): %d", t.Spl5)
	writeIndentfln(b, 3, "SVN 6 (reserved): %d", t.Spl6)
	writeIndentfln(b, 3, "SVN 7 (reserved): %d", t.Spl7)
	writeIndentfln(b, 3, "SEV-SNP firmware SVN: %d", t.SNP)
	writeIndentfln(b, 3, "Microcode SVN: %d", t.Microcode)
}

// newTCBVersion creates a TCB version from a kds.TCBVersion.
func newTCBVersion(tcbVersion kds.TCBVersion) TCBVersion {
	tcb := kds.DecomposeTCBVersion(tcbVersion)
	return TCBVersion{
		Bootloader: tcb.BlSpl,
		TEE:        tcb.TeeSpl,
		SNP:        tcb.SnpSpl,
		Microcode:  tcb.UcodeSpl,
		Spl4:       tcb.Spl4,
		Spl5:       tcb.Spl5,
		Spl6:       tcb.Spl6,
		Spl7:       tcb.Spl7,
	}
}

// PlatformInfo contains the platform information.
type PlatformInfo struct {
	SMT  bool `json:"smt"`
	TSME bool `json:"tsme"`
}

// SignerInfo contains the signer information.
type SignerInfo struct {
	AuthorKey   bool   `json:"author_key_en"`
	MaskChipKey bool   `json:"mask_chip_key"`
	SigningKey  string `json:"signing_key"`
}

// SNPReport contains the SNP report data.
type SNPReport struct {
	Version              uint32       `json:"version"`
	GuestSvn             uint32       `json:"guest_svn"`
	PolicyABIMinor       uint8        `json:"policy_abi_minor"`
	PolicyABIMajor       uint8        `json:"policy_abi_major"`
	PolicySMT            bool         `json:"policy_symmetric_multi_threading"`
	PolicyMigrationAgent bool         `json:"policy_migration_agent"`
	PolicyDebug          bool         `json:"policy_debug"`
	PolicySingleSocket   bool         `json:"policy_single_socket"`
	FamilyID             []byte       `json:"family_id"`
	ImageID              []byte       `json:"image_id"`
	Vmpl                 uint32       `json:"vmpl"`
	SignatureAlgo        uint32       `json:"signature_algo"`
	CurrentTCB           TCBVersion   `json:"current_tcb"`
	PlatformInfo         PlatformInfo `json:"platform_info"`
	SignerInfo           SignerInfo   `json:"signer_info"`
	ReportData           []byte       `json:"report_data"`
	Measurement          []byte       `json:"measurement"`
	HostData             []byte       `json:"host_data"`
	IDKeyDigest          []byte       `json:"id_key_digest"`
	AuthorKeyDigest      []byte       `json:"author_key_digest"`
	ReportID             []byte       `json:"report_id"`
	ReportIDMa           []byte       `json:"report_id_ma"`
	ReportedTCB          TCBVersion   `json:"reported_tcb"`
	ChipID               []byte       `json:"chip_id"`
	CommittedTCB         TCBVersion   `json:"committed_tcb"`
	CurrentBuild         uint32       `json:"current_build"`
	CurrentMinor         uint32       `json:"current_minor"`
	CurrentMajor         uint32       `json:"current_major"`
	CommittedBuild       uint32       `json:"committed_build"`
	CommittedMinor       uint32       `json:"committed_minor"`
	CommittedMajor       uint32       `json:"committed_major"`
	LaunchTCB            TCBVersion   `json:"launch_tcb"`
	Signature            []byte       `json:"signature"`
}

// newSNPReport parses a marshalled SNP report and returns a SNPReport object.
func newSNPReport(reportBytes []byte) (SNPReport, error) {
	report, err := abi.ReportToProto(reportBytes)
	if err != nil {
		return SNPReport{}, fmt.Errorf("parsing report to proto: %w", err)
	}

	policy, err := abi.ParseSnpPolicy(report.Policy)
	if err != nil {
		return SNPReport{}, fmt.Errorf("parsing policy: %w", err)
	}

	platformInfo, err := abi.ParseSnpPlatformInfo(report.PlatformInfo)
	if err != nil {
		return SNPReport{}, fmt.Errorf("parsing platform info: %w", err)
	}

	signature, err := abi.ReportToSignatureDER(reportBytes)
	if err != nil {
		return SNPReport{}, fmt.Errorf("parsing signature: %w", err)
	}

	signerInfo, err := abi.ParseSignerInfo(report.SignerInfo)
	if err != nil {
		return SNPReport{}, fmt.Errorf("parsing signer info: %w", err)
	}
	return SNPReport{
		Version:              report.Version,
		GuestSvn:             report.GuestSvn,
		PolicyABIMinor:       policy.ABIMinor,
		PolicyABIMajor:       policy.ABIMajor,
		PolicySMT:            policy.SMT,
		PolicyMigrationAgent: policy.MigrateMA,
		PolicyDebug:          policy.Debug,
		PolicySingleSocket:   policy.SingleSocket,
		FamilyID:             report.FamilyId,
		ImageID:              report.ImageId,
		Vmpl:                 report.Vmpl,
		SignatureAlgo:        report.SignatureAlgo,
		CurrentTCB:           newTCBVersion(kds.TCBVersion(report.CurrentTcb)),
		PlatformInfo: PlatformInfo{
			SMT:  platformInfo.SMTEnabled,
			TSME: platformInfo.TSMEEnabled,
		},
		SignerInfo: SignerInfo{
			AuthorKey:   signerInfo.AuthorKeyEn,
			MaskChipKey: signerInfo.MaskChipKey,
			SigningKey:  signerInfo.SigningKey.String(),
		},
		ReportData:      report.ReportData,
		Measurement:     report.Measurement,
		HostData:        report.HostData,
		IDKeyDigest:     report.IdKeyDigest,
		AuthorKeyDigest: report.AuthorKeyDigest,
		ReportID:        report.ReportId,
		ReportIDMa:      report.ReportIdMa,
		ReportedTCB:     newTCBVersion(kds.TCBVersion(report.ReportedTcb)),
		ChipID:          report.ChipId,
		CommittedTCB:    newTCBVersion(kds.TCBVersion(report.CommittedTcb)),
		CurrentBuild:    report.CurrentBuild,
		CurrentMinor:    report.CurrentMinor,
		CurrentMajor:    report.CurrentMajor,
		CommittedBuild:  report.CommittedBuild,
		CommittedMinor:  report.CommittedMinor,
		CommittedMajor:  report.CommittedMajor,
		LaunchTCB:       newTCBVersion(kds.TCBVersion(report.LaunchTcb)),
		Signature:       signature,
	}, nil
}

// formatString builds a string representation of a SNP report that is inteded for console output.
func (s *SNPReport) formatString(b *strings.Builder) {
	writeIndentfln(b, 1, "SNP Report:")
	writeIndentfln(b, 2, "Version: %d", s.Version)
	writeIndentfln(b, 2, "Guest SVN: %d", s.GuestSvn)
	writeIndentfln(b, 2, "Policy:")
	writeIndentfln(b, 3, "ABI Minor: %d", s.PolicyABIMinor)
	writeIndentfln(b, 3, "ABI Major: %d", s.PolicyABIMajor)
	writeIndentfln(b, 3, "Symmetric Multithreading enabled: %t", s.PolicySMT)
	writeIndentfln(b, 3, "Migration agent enabled: %t", s.PolicyMigrationAgent)
	writeIndentfln(b, 3, "Debugging enabled (host decryption of VM): %t", s.PolicyDebug)
	writeIndentfln(b, 3, "Single socket enabled: %t", s.PolicySingleSocket)
	writeIndentfln(b, 2, "Family ID: %x", s.FamilyID)
	writeIndentfln(b, 2, "Image ID: %x", s.ImageID)
	writeIndentfln(b, 2, "VMPL: %d", s.Vmpl)
	writeIndentfln(b, 2, "Signature Algorithm: %d", s.SignatureAlgo)
	writeIndentfln(b, 2, "Current TCB:")
	s.CurrentTCB.formatString(b)
	writeIndentfln(b, 2, "Platform Info:")
	writeIndentfln(b, 3, "Symmetric Multithreading enabled (SMT): %t", s.PlatformInfo.SMT)
	writeIndentfln(b, 3, "Transparent secure memory encryption (TSME): %t", s.PlatformInfo.TSME)
	writeIndentfln(b, 2, "Signer Info:")
	writeIndentfln(b, 3, "Author Key Enabled: %t", s.SignerInfo.AuthorKey)
	writeIndentfln(b, 3, "Chip ID Masking: %t", s.SignerInfo.MaskChipKey)
	writeIndentfln(b, 3, "Signing Type: %s", s.SignerInfo.SigningKey)
	writeIndentfln(b, 2, "Report Data: %x", s.ReportData)
	writeIndentfln(b, 2, "Measurement: %x", s.Measurement)
	writeIndentfln(b, 2, "Host Data: %x", s.HostData)
	writeIndentfln(b, 2, "ID Key Digest: %x", s.IDKeyDigest)
	writeIndentfln(b, 2, "Author Key Digest: %x", s.AuthorKeyDigest)
	writeIndentfln(b, 2, "Report ID: %x", s.ReportID)
	writeIndentfln(b, 2, "Report ID MA: %x", s.ReportIDMa)
	writeIndentfln(b, 2, "Reported TCB:")
	s.ReportedTCB.formatString(b)
	writeIndentfln(b, 2, "Chip ID: %x", s.ChipID)
	writeIndentfln(b, 2, "Committed TCB:")
	s.CommittedTCB.formatString(b)
	writeIndentfln(b, 2, "Current Build: %d", s.CurrentBuild)
	writeIndentfln(b, 2, "Current Minor: %d", s.CurrentMinor)
	writeIndentfln(b, 2, "Current Major: %d", s.CurrentMajor)
	writeIndentfln(b, 2, "Committed Build: %d", s.CommittedBuild)
	writeIndentfln(b, 2, "Committed Minor: %d", s.CommittedMinor)
	writeIndentfln(b, 2, "Committed Major: %d", s.CommittedMajor)
	writeIndentfln(b, 2, "Launch TCB:")
	s.LaunchTCB.formatString(b)
	writeIndentfln(b, 2, "Signature (DER):")
	writeIndentfln(b, 3, "%x", s.Signature)
}

// MaaTokenClaims contains the MAA token claims.
type MaaTokenClaims struct {
	jwt.RegisteredClaims
	Secureboot                               bool   `json:"secureboot,omitempty"`
	XMsAttestationType                       string `json:"x-ms-attestation-type,omitempty"`
	XMsAzurevmAttestationProtocolVer         string `json:"x-ms-azurevm-attestation-protocol-ver,omitempty"`
	XMsAzurevmAttestedPcrs                   []int  `json:"x-ms-azurevm-attested-pcrs,omitempty"`
	XMsAzurevmBootdebugEnabled               bool   `json:"x-ms-azurevm-bootdebug-enabled,omitempty"`
	XMsAzurevmDbvalidated                    bool   `json:"x-ms-azurevm-dbvalidated,omitempty"`
	XMsAzurevmDbxvalidated                   bool   `json:"x-ms-azurevm-dbxvalidated,omitempty"`
	XMsAzurevmDebuggersdisabled              bool   `json:"x-ms-azurevm-debuggersdisabled,omitempty"`
	XMsAzurevmDefaultSecurebootkeysvalidated bool   `json:"x-ms-azurevm-default-securebootkeysvalidated,omitempty"`
	XMsAzurevmElamEnabled                    bool   `json:"x-ms-azurevm-elam-enabled,omitempty"`
	XMsAzurevmFlightsigningEnabled           bool   `json:"x-ms-azurevm-flightsigning-enabled,omitempty"`
	XMsAzurevmHvciPolicy                     int    `json:"x-ms-azurevm-hvci-policy,omitempty"`
	XMsAzurevmHypervisordebugEnabled         bool   `json:"x-ms-azurevm-hypervisordebug-enabled,omitempty"`
	XMsAzurevmIsWindows                      bool   `json:"x-ms-azurevm-is-windows,omitempty"`
	XMsAzurevmKerneldebugEnabled             bool   `json:"x-ms-azurevm-kerneldebug-enabled,omitempty"`
	XMsAzurevmOsbuild                        string `json:"x-ms-azurevm-osbuild,omitempty"`
	XMsAzurevmOsdistro                       string `json:"x-ms-azurevm-osdistro,omitempty"`
	XMsAzurevmOstype                         string `json:"x-ms-azurevm-ostype,omitempty"`
	XMsAzurevmOsversionMajor                 int    `json:"x-ms-azurevm-osversion-major,omitempty"`
	XMsAzurevmOsversionMinor                 int    `json:"x-ms-azurevm-osversion-minor,omitempty"`
	XMsAzurevmSigningdisabled                bool   `json:"x-ms-azurevm-signingdisabled,omitempty"`
	XMsAzurevmTestsigningEnabled             bool   `json:"x-ms-azurevm-testsigning-enabled,omitempty"`
	XMsAzurevmVmid                           string `json:"x-ms-azurevm-vmid,omitempty"`
	XMsIsolationTee                          struct {
		XMsAttestationType  string `json:"x-ms-attestation-type,omitempty"`
		XMsComplianceStatus string `json:"x-ms-compliance-status,omitempty"`
		XMsRuntime          struct {
			Keys []struct {
				E      string   `json:"e,omitempty"`
				KeyOps []string `json:"key_ops,omitempty"`
				Kid    string   `json:"kid,omitempty"`
				Kty    string   `json:"kty,omitempty"`
				N      string   `json:"n,omitempty"`
			} `json:"keys,omitempty"`
			VMConfiguration struct {
				ConsoleEnabled bool   `json:"console-enabled,omitempty"`
				CurrentTime    int    `json:"current-time,omitempty"`
				SecureBoot     bool   `json:"secure-boot,omitempty"`
				TpmEnabled     bool   `json:"tpm-enabled,omitempty"`
				VMUniqueID     string `json:"vmUniqueId,omitempty"`
			} `json:"vm-configuration,omitempty"`
		} `json:"x-ms-runtime,omitempty"`
		XMsSevsnpvmAuthorkeydigest   string `json:"x-ms-sevsnpvm-authorkeydigest,omitempty"`
		XMsSevsnpvmBootloaderSvn     int    `json:"x-ms-sevsnpvm-bootloader-svn,omitempty"`
		XMsSevsnpvmFamilyID          string `json:"x-ms-sevsnpvm-familyId,omitempty"`
		XMsSevsnpvmGuestsvn          int    `json:"x-ms-sevsnpvm-guestsvn,omitempty"`
		XMsSevsnpvmHostdata          string `json:"x-ms-sevsnpvm-hostdata,omitempty"`
		XMsSevsnpvmIdkeydigest       string `json:"x-ms-sevsnpvm-idkeydigest,omitempty"`
		XMsSevsnpvmImageID           string `json:"x-ms-sevsnpvm-imageId,omitempty"`
		XMsSevsnpvmIsDebuggable      bool   `json:"x-ms-sevsnpvm-is-debuggable,omitempty"`
		XMsSevsnpvmLaunchmeasurement string `json:"x-ms-sevsnpvm-launchmeasurement,omitempty"`
		XMsSevsnpvmMicrocodeSvn      int    `json:"x-ms-sevsnpvm-microcode-svn,omitempty"`
		XMsSevsnpvmMigrationAllowed  bool   `json:"x-ms-sevsnpvm-migration-allowed,omitempty"`
		XMsSevsnpvmReportdata        string `json:"x-ms-sevsnpvm-reportdata,omitempty"`
		XMsSevsnpvmReportid          string `json:"x-ms-sevsnpvm-reportid,omitempty"`
		XMsSevsnpvmSmtAllowed        bool   `json:"x-ms-sevsnpvm-smt-allowed,omitempty"`
		XMsSevsnpvmSnpfwSvn          int    `json:"x-ms-sevsnpvm-snpfw-svn,omitempty"`
		XMsSevsnpvmTeeSvn            int    `json:"x-ms-sevsnpvm-tee-svn,omitempty"`
		XMsSevsnpvmVmpl              int    `json:"x-ms-sevsnpvm-vmpl,omitempty"`
	} `json:"x-ms-isolation-tee,omitempty"`
	XMsPolicyHash string `json:"x-ms-policy-hash,omitempty"`
	XMsRuntime    struct {
		ClientPayload struct {
			Nonce string `json:"nonce,omitempty"`
		} `json:"client-payload,omitempty"`
		Keys []struct {
			E      string   `json:"e,omitempty"`
			KeyOps []string `json:"key_ops,omitempty"`
			Kid    string   `json:"kid,omitempty"`
			Kty    string   `json:"kty,omitempty"`
			N      string   `json:"n,omitempty"`
		} `json:"keys,omitempty"`
	} `json:"x-ms-runtime,omitempty"`
	XMsVer string `json:"x-ms-ver,omitempty"`
}

// newMAAToken parses a MAA token and returns a MaaTokenClaims object.
func newMAAToken(ctx context.Context, rawToken, attestationServiceURL string) (MaaTokenClaims, error) {
	var claims MaaTokenClaims
	_, err := jwt.ParseWithClaims(rawToken, &claims, keyFromJKUFunc(ctx, attestationServiceURL), jwt.WithIssuedAt())
	return claims, err
}

// formatString builds a string representation of a MAA token that is inteded for console output.
func (m *MaaTokenClaims) formatString(b *strings.Builder) error {
	out, err := json.MarshalIndent(m, "\t\t", "  ")
	if err != nil {
		return fmt.Errorf("marshaling claims: %w", err)
	}

	b.WriteString("\tMicrosoft Azure Attestation Token:\n\t")
	b.WriteString(string(out))

	return nil
}

// writeIndentfln writes a formatted string to the builder with the given indentation level
// and a newline at the end.
func writeIndentfln(b *strings.Builder, indentLvl int, format string, args ...any) {
	for i := 0; i < indentLvl; i++ {
		b.WriteByte('\t')
	}
	b.WriteString(fmt.Sprintf(format+"\n", args...))
}

// keyFromJKUFunc returns a function that gets the JSON Web Key URI from the token
// and fetches the key from that URI. The keys are then parsed, and the key with
// the kid that matches the token header is returned.
func keyFromJKUFunc(ctx context.Context, webKeysURLBase string) func(token *jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		webKeysURL, err := url.JoinPath(webKeysURLBase, "certs")
		if err != nil {
			return nil, fmt.Errorf("joining web keys base URL with path: %w", err)
		}

		if token.Header["alg"] != "RS256" {
			return nil, fmt.Errorf("invalid signing algorithm: %s", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid kid: %v", token.Header["kid"])
		}
		jku, ok := token.Header["jku"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid jku: %v", token.Header["jku"])
		}
		if jku != webKeysURL {
			return nil, fmt.Errorf("jku from token (%s) does not match configured attestation service (%s)", jku, webKeysURL)
		}

		keySetBytes, err := httpGet(ctx, jku)
		if err != nil {
			return nil, fmt.Errorf("getting signing keys from jku %s: %w", jku, err)
		}

		var rawKeySet struct {
			Keys []struct {
				X5c [][]byte
				Kid string
			}
		}

		if err := json.Unmarshal(keySetBytes, &rawKeySet); err != nil {
			return nil, err
		}

		for _, key := range rawKeySet.Keys {
			if key.Kid != kid {
				continue
			}
			cert, err := x509.ParseCertificate(key.X5c[0])
			if err != nil {
				return nil, fmt.Errorf("parsing certificate: %w", err)
			}

			return cert.PublicKey, nil
		}

		return nil, fmt.Errorf("no key found for kid %s", kid)
	}
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type debugLog interface {
	Debug(format string, args ...any)
}
