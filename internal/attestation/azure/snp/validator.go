/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	spb "github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/validate"
	"github.com/google/go-sev-guest/verify"
	"github.com/google/go-sev-guest/verify/trust"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
)

// Validator for Azure confidential VM attestation.
type Validator struct {
	variant.AzureSEVSNP
	*vtpm.Validator
	hclValidator hclAkValidator
	maa          maaValidator
	getter       trust.HTTPSGetter

	attestationVerifier  attestationVerifier
	attestationValidator attestationValidator

	// ASK certificate cached by the join service.
	cachedASKCert *x509.Certificate

	config *config.AzureSEVSNP

	log attestation.Logger
}

type attestationVerifier interface {
	SNPAttestation(attestation *spb.Attestation, options *verify.Options) error
}

type attestationValidator interface {
	SNPAttestation(attestation *spb.Attestation, options *validate.Options) error
}

type attestationVerifierImpl struct{}

// SNPAttestation verifies the report signature, the VCEK certificate, as well as the certificate chain of the attestation report.
func (attestationVerifierImpl) SNPAttestation(attestation *spb.Attestation, options *verify.Options) error {
	return verify.SnpAttestation(attestation, options)
}

type attestationValidatorImpl struct{}

// SNPAttestation validates the attestation report against the given set of constraints.
func (attestationValidatorImpl) SNPAttestation(attestation *spb.Attestation, options *validate.Options) error {
	return validate.SnpAttestation(attestation, options)
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(cfg *config.AzureSEVSNP, log attestation.Logger) *Validator {
	if log == nil {
		log = nopAttestationLogger{}
	}
	v := &Validator{
		hclValidator:         &azureInstanceInfo{},
		maa:                  newMAAClient(),
		config:               cfg,
		log:                  log,
		getter:               trust.DefaultHTTPSGetter(),
		attestationVerifier:  attestationVerifierImpl{},
		attestationValidator: attestationValidatorImpl{},
	}
	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		v.getTrustedKey,
		// stub, since SEV-SNP attestation is already verified in trustedKeyFromSNP().
		func(vtpm.AttestationDocument, *attest.MachineState) error {
			return nil
		},
		log,
	)
	return v
}

// WithCachedASKCert sets the cached ASK certificate.
func (v *Validator) WithCachedASKCert(cert *x509.Certificate) *Validator {
	v.cachedASKCert = cert
	return v
}

// getTrustedKey establishes trust in the given public key.
// It does so by verifying the SNP attestation document.
func (v *Validator) getTrustedKey(ctx context.Context, attDoc vtpm.AttestationDocument, extraData []byte) (crypto.PublicKey, error) {
	// ARK, specified in Constellation config.
	trustedArk := x509.Certificate(v.config.AMDRootKey)

	// fallback certificates, used if not present in THIM response.
	cachedCerts := sevSnpCerts{
		ask: v.cachedASKCert,
		ark: &trustedArk,
	}

	// transform the instanceInfo received from Microsoft into a verifiable attestation report format.
	var instanceInfo azureInstanceInfo
	if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
		return nil, fmt.Errorf("unmarshalling instanceInfo: %w", err)
	}

	att, err := instanceInfo.attestationWithCerts(v.log, v.getter, cachedCerts)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation report: %w", err)
	}

	// Verify the attestation report's certificates.
	// ASK, as cached in joinservice or reported from THIM / KDS.
	ask, err := x509.ParseCertificate(att.CertificateChain.AskCert)
	if err != nil {
		return nil, fmt.Errorf("parsing ASK certificate: %w", err)
	}

	verifyOpts := &verify.Options{
		TrustedRoots: map[string][]*trust.AMDRootCerts{
			"Milan": {
				{
					Product: "Milan",
					ProductCerts: &trust.ProductCerts{
						Ask: ask,
						Ark: &trustedArk,
					},
				},
			},
		},
	}
	if err := v.attestationVerifier.SNPAttestation(att, verifyOpts); err != nil {
		return nil, fmt.Errorf("verifying SNP attestation: %w", err)
	}

	// Checks if the attestation report matches the given constraints.
	// Some constraints are implicitly checked by validate.SnpAttestation:
	// - the report is not expired
	if err := v.attestationValidator.SNPAttestation(att, &validate.Options{
		GuestPolicy: abi.SnpPolicy{
			Debug: false, // Debug means the VM can be decrypted by the host for debugging purposes and thus is not allowed.
			SMT:   true,  // Allow Simultaneous Multi-Threading (SMT). Normally, we would want to disable SMT
			// but Azure does not allow to disable it.
		},
		VMPL: new(int), // Checks that Virtual Machine Privilege Level (VMPL) is 0.
		// This checks that the reported TCB version is equal or greater than the minimum specified in the config.
		MinimumTCB: kds.TCBParts{
			BlSpl:    v.config.BootloaderVersion.Value, // Bootloader
			TeeSpl:   v.config.TEEVersion.Value,        // TEE (Secure OS)
			SnpSpl:   v.config.SNPVersion.Value,        // SNP
			UcodeSpl: v.config.MicrocodeVersion.Value,  // Microcode
		},
		// This checks that the reported LaunchTCB version is equal or greater than the minimum specified in the config.
		MinimumLaunchTCB: kds.TCBParts{
			BlSpl:    v.config.BootloaderVersion.Value, // Bootloader
			TeeSpl:   v.config.TEEVersion.Value,        // TEE (Secure OS)
			SnpSpl:   v.config.SNPVersion.Value,        // SNP
			UcodeSpl: v.config.MicrocodeVersion.Value,  // Microcode
		},
		// Check that CurrentTCB >= CommittedTCB.
		PermitProvisionalFirmware: true,
		// Check if the IDKey hash in the report is in the list of accepted hashes.
		TrustedIDKeyHashes: v.config.FirmwareSignerConfig.AcceptedKeyDigests,
		// The IDKey hash should not be checked if the enforcement policy is set to MAAFallback or WarnOnly to prevent
		// an error from being returned because of the TrustedIDKeyHashes validation. In this case, we should perform a
		// custom check of the MAA-specific values later. Right now, this is a double check, since a custom MAA check
		// is performed either way.
		RequireIDBlock: v.config.FirmwareSignerConfig.EnforcementPolicy == idkeydigest.Equal,
	}); err != nil {
		return nil, fmt.Errorf("validating SNP attestation: %w", err)
	}
	// Custom check of the IDKeyDigests, taking care of the WarnOnly / MAAFallback cases,
	// but also double-checking the IDKeyDigests if the enforcement policy is set to Equal.
	if err := v.checkIDKeyDigest(ctx, att, instanceInfo.MAAToken, extraData); err != nil {
		return nil, fmt.Errorf("checking IDKey digests: %w", err)
	}

	// Decode the public area of the attestation key and validate its trustworthiness.
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, err
	}
	if err = v.hclValidator.validateAk(instanceInfo.RuntimeData, att.Report.ReportData, pubArea.RSAParameters); err != nil {
		return nil, fmt.Errorf("validating HCLAkPub: %w", err)
	}

	return pubArea.Key()
}

// checkIDKeyDigest validates the IDKeyDigest in the given attestation report against the accepted IDKeyDigests in the
// validator's config. If an IDKeyDigest is present in the report that is not in the accepted IDKeyDigests, the validation proceeds
// according to the enforcement policy. If the enforcement policy is set to MAAFallback, the maaToken is validated against the MAA.
// If the enforcement policy is set to WarnOnly, a warning is logged. If the enforcement policy is set to neither WarnOnly or MAAFallback, an
// error is returned.
func (v *Validator) checkIDKeyDigest(ctx context.Context, report *spb.Attestation, maaToken string, extraData []byte) error {
	for _, digest := range v.config.FirmwareSignerConfig.AcceptedKeyDigests {
		if bytes.Equal(digest, report.Report.IdKeyDigest) {
			return nil
		}
	}

	// IDKeyDigest that was not expected is present, check the enforcement policy and verify against
	// the MAA if necessary.
	switch v.config.FirmwareSignerConfig.EnforcementPolicy {
	case idkeydigest.MAAFallback:
		v.log.Infof(
			"Configured idkeydigests %x don't contain reported idkeydigest %x, falling back to MAA validation",
			v.config.FirmwareSignerConfig.AcceptedKeyDigests,
			report.Report.IdKeyDigest,
		)
		return v.maa.validateToken(ctx, v.config.FirmwareSignerConfig.MAAURL, maaToken, extraData)
	case idkeydigest.WarnOnly:
		v.log.Warnf(
			"Configured idkeydigests %x don't contain reported idkeydigest %x",
			v.config.FirmwareSignerConfig.AcceptedKeyDigests,
			report.Report.IdKeyDigest,
		)
	default:
		return fmt.Errorf(
			"configured idkeydigests %x don't contain reported idkeydigest %x",
			v.config.FirmwareSignerConfig.AcceptedKeyDigests,
			report.Report.IdKeyDigest,
		)
	}

	// No IDKeyDigest that was not expected is present.
	return nil
}

// azureInstanceInfo contains the necessary information to establish trust in
// an Azure CVM.
type azureInstanceInfo struct {
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

// attestationWithCerts returns a formatted version of the attestation report and its certificates from the instanceInfo.
// Certificates are retrieved in the following precedence:
// 1. ASK or ARK from THIM
// 2. ASK or ARK from fallbackCerts
// 3. ASK or ARK from AMD KDS.
func (a *azureInstanceInfo) attestationWithCerts(logger attestation.Logger, getter trust.HTTPSGetter,
	fallbackCerts sevSnpCerts,
) (*spb.Attestation, error) {
	report, err := abi.ReportToProto(a.AttestationReport)
	if err != nil {
		return nil, fmt.Errorf("converting report to proto: %w", err)
	}

	// Product info as reported through CPUID[EAX=1]
	sevProduct := abi.DefaultSevProduct()
	productName := kds.ProductString(sevProduct)

	att := &spb.Attestation{
		Report:           report,
		CertificateChain: &spb.CertificateChain{},
		Product:          sevProduct,
	}

	// If the VCEK certificate is present, parse it and format it.
	vcek, err := a.parseVCEK()
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
	ask, ark, err := a.parseCertChain()
	if err != nil {
		logger.Warnf("Error parsing certificate chain: %v", err)
	}
	if ask != nil {
		att.CertificateChain.AskCert = ask.Raw
	}
	if ark != nil {
		att.CertificateChain.ArkCert = ark.Raw
	}
	// If a cached ASK or an ARK specified by the Constellation config is present, use it.
	if fallbackCerts.ask != nil {
		att.CertificateChain.AskCert = fallbackCerts.ask.Raw
	}
	if fallbackCerts.ark != nil {
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
		if att.CertificateChain.AskCert == nil {
			att.CertificateChain.AskCert = kdsCertChain.Ask.Raw
		}
		if att.CertificateChain.ArkCert == nil {
			att.CertificateChain.ArkCert = kdsCertChain.Ark.Raw
		}
	}

	return att, nil
}

type sevSnpCerts struct {
	ask *x509.Certificate
	ark *x509.Certificate
}

// parseCertChain parses the certificate chain from the instanceInfo into x509-formatted ASK and ARK certificates.
// If less than 2 certificates are present, only the present certificate is returned.
// If more than 2 certificates are present, an error is returned.
func (a *azureInstanceInfo) parseCertChain() (ask, ark *x509.Certificate, retErr error) {
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

// parseVCEK parses the VCEK certificate from the instanceInfo into an x509-formatted certificate.
// If the VCEK certificate is not present, nil is returned.
func (a *azureInstanceInfo) parseVCEK() (*x509.Certificate, error) {
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

// validateAk validates that the attestation key from the TPM is trustworthy. The steps are:
// 1. runtime data read from the TPM has the same sha256 digest as reported in `report_data` of the SNP report.
// 2. modulus reported in runtime data matches modulus from key at idx 0x81000003.
// 3. exponent reported in runtime data matches exponent from key at idx 0x81000003.
// The function is currently tested manually on a Azure Ubuntu CVM.
func (a *azureInstanceInfo) validateAk(runtimeDataRaw []byte, reportData []byte, rsaParameters *tpm2.RSAParams) error {
	var runtimeData runtimeData
	if err := json.Unmarshal(runtimeDataRaw, &runtimeData); err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}

	sum := sha256.Sum256(runtimeDataRaw)
	if len(reportData) < len(sum) {
		return fmt.Errorf("reportData has unexpected size: %d", len(reportData))
	}
	if !bytes.Equal(sum[:], reportData[:len(sum)]) {
		return errors.New("unexpected runtimeData digest in TPM")
	}

	if len(runtimeData.Keys) < 1 {
		return errors.New("did not receive any keys in runtime data")
	}
	rawN, err := base64.RawURLEncoding.DecodeString(runtimeData.Keys[0].N)
	if err != nil {
		return fmt.Errorf("decoding modulus string: %w", err)
	}
	if !bytes.Equal(rawN, rsaParameters.ModulusRaw) {
		return fmt.Errorf("unexpected modulus value in TPM")
	}

	rawE, err := base64.RawURLEncoding.DecodeString(runtimeData.Keys[0].E)
	if err != nil {
		return fmt.Errorf("decoding exponent string: %w", err)
	}
	paddedRawE := make([]byte, 4)
	copy(paddedRawE, rawE)
	exponent := binary.LittleEndian.Uint32(paddedRawE)

	// According to this comment [1] the TPM uses "0" to represent the default exponent "65537".
	// The go tpm library also reports the exponent as 0. Thus we have to handle it specially.
	// [1] https://github.com/tpm2-software/tpm2-tools/pull/1973#issue-596685005
	if !((exponent == 65537 && rsaParameters.ExponentRaw == 0) || exponent == rsaParameters.ExponentRaw) {
		return fmt.Errorf("unexpected N value in TPM")
	}

	return nil
}

// hclAkValidator validates an attestation key issued by the Host Compatibility Layer (HCL).
// The HCL is written by Azure, and sits between the Hypervisor and CVM OS.
// The HCL runs in the protected context of the CVM.
type hclAkValidator interface {
	validateAk(runtimeDataRaw []byte, reportData []byte, rsaParameters *tpm2.RSAParams) error
}

// akPub are the public parameters of an RSA attestation key.
type akPub struct {
	E string
	N string
}

type runtimeData struct {
	Keys []akPub
}

// nopAttestationLogger is a no-op implementation of AttestationLogger.
type nopAttestationLogger struct{}

// Infof is a no-op.
func (nopAttestationLogger) Infof(string, ...interface{}) {}

// Warnf is a no-op.
func (nopAttestationLogger) Warnf(string, ...interface{}) {}

type maaValidator interface {
	validateToken(ctx context.Context, maaURL string, token string, extraData []byte) error
}
