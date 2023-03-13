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
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	internalCrypto "github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/tpm2"
)

const (
	// AMD root key. Received from the AMD Key Distribution System API (KDS).
	arkPEM            = "-----BEGIN CERTIFICATE-----\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\nAFZEAwoKCQ==\n-----END CERTIFICATE-----\n"
	bootloaderVersion = 2
	teeVersion        = 0
	snpVersion        = 6
	microcodeVersion  = 93
)

// Validator for Azure confidential VM attestation.
type Validator struct {
	oid.AzureSEVSNP
	*vtpm.Validator
	hclValidator hclAkValidator
	maa          maaValidator

	idKeyDigests       idkeydigest.IDKeyDigests
	enforceIDKeyDigest idkeydigest.EnforceIDKeyDigest
	maaURL             string

	log vtpm.AttestationLogger
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(pcrs measurements.M, idKeyConf idkeydigest.Config, log vtpm.AttestationLogger) *Validator {
	if log == nil {
		log = nopAttestationLogger{}
	}
	v := &Validator{
		hclValidator:       &azureInstanceInfo{},
		maa:                newMAAClient(),
		idKeyDigests:       idKeyConf.IDKeyDigests,
		enforceIDKeyDigest: idKeyConf.EnforcementPolicy,
		maaURL:             idKeyConf.MAAURL,
		log:                log,
	}
	v.Validator = vtpm.NewValidator(
		pcrs,
		v.getTrustedKey,
		validateCVM,
		log,
	)
	return v
}

// validateCVM is a stub, since SEV-SNP attestation is already verified in trustedKeyFromSNP().
func validateCVM(vtpm.AttestationDocument, *attest.MachineState) error {
	return nil
}

func newSNPReportFromBytes(reportRaw []byte) (snpAttestationReport, error) {
	var report snpAttestationReport
	if err := binary.Read(bytes.NewReader(reportRaw), binary.LittleEndian, &report); err != nil {
		return snpAttestationReport{}, fmt.Errorf("reading attestation report: %w", err)
	}

	return report, nil
}

func reverseEndian(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
}

// getTrustedKey establishes trust in the given public key.
// It does so by verifying the SNP attestation statement in instanceInfo.
func (v *Validator) getTrustedKey(ctx context.Context, attDoc vtpm.AttestationDocument, extraData []byte) (crypto.PublicKey, error) {
	var instanceInfo azureInstanceInfo
	if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
		return nil, fmt.Errorf("unmarshalling instanceInfoRaw: %w", err)
	}

	report, err := newSNPReportFromBytes(instanceInfo.AttestationReport)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation report: %w", err)
	}

	vcek, err := validateVCEK(instanceInfo.Vcek, instanceInfo.CertChain)
	if err != nil {
		return nil, fmt.Errorf("validating VCEK: %w", err)
	}

	if err := v.validateSNPReport(ctx, vcek, report, instanceInfo.MAAToken, extraData); err != nil {
		return nil, fmt.Errorf("validating SNP report: %w", err)
	}

	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, err
	}

	if err = v.hclValidator.validateAk(instanceInfo.RuntimeData, report.ReportData[:], pubArea.RSAParameters); err != nil {
		return nil, fmt.Errorf("validating HCLAkPub: %w", err)
	}

	return pubArea.Key()
}

// validateVCEK takes the PEM-encoded X509 certificate VCEK, ASK and ARK and verifies the integrity of the chain.
// ARK (hardcoded) validates ASK (cloud metadata API) validates VCEK (cloud metadata API).
func validateVCEK(vcekRaw []byte, certChain []byte) (*x509.Certificate, error) {
	vcek, err := internalCrypto.PemToX509Cert(vcekRaw)
	if err != nil {
		return nil, fmt.Errorf("loading vcek: %w", err)
	}

	ark, err := internalCrypto.PemToX509Cert([]byte(arkPEM))
	if err != nil {
		return nil, fmt.Errorf("loading arkPEM: %w", err)
	}

	// certChain includes two PEM encoded certs. The ASK and the ARK, in that order.
	ask, err := internalCrypto.PemToX509Cert(certChain)
	if err != nil {
		return nil, fmt.Errorf("loading askPEM: %w", err)
	}

	if err = ask.CheckSignatureFrom(ark); err != nil {
		return nil, &askError{err}
	}

	if err = vcek.CheckSignatureFrom(ask); err != nil {
		return nil, &vcekError{err}
	}

	return vcek, nil
}

func (v *Validator) validateSNPReport(
	ctx context.Context, cert *x509.Certificate, report snpAttestationReport, maaToken string, extraData []byte,
) error {
	if report.Policy.Debug() {
		return errDebugEnabled
	}

	if !report.CommittedTCB.isVersion(bootloaderVersion, teeVersion, snpVersion, microcodeVersion) {
		return &versionError{"COMMITTED_TCB", report.CommittedTCB}
	}
	if report.LaunchTCB != report.CommittedTCB {
		return &versionError{"LAUNCH_TCB", report.LaunchTCB}
	}
	if !report.CommittedTCB.supersededBy(report.CurrentTCB) {
		return &versionError{"CURRENT_TCB", report.CurrentTCB}
	}

	if err := validateVCEKExtensions(cert, report); err != nil {
		return fmt.Errorf("mismatching vcek extensions: %w", err)
	}

	sigR := report.Signature.R[:]
	sigS := report.Signature.S[:]

	// Table 107 in https://www.amd.com/system/files/TechDocs/56860.pdf mentions little endian signature components.
	// They come out of the certificate as big endian.
	reverseEndian(sigR)
	reverseEndian(sigS)

	rParam := new(big.Int).SetBytes(sigR)
	sParam := new(big.Int).SetBytes(sigS)
	sequence := ecdsaSig{rParam, sParam}
	sigEncoded, err := asn1.Marshal(sequence)
	if err != nil {
		return fmt.Errorf("marshalling ecdsa signature: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, report); err != nil {
		return fmt.Errorf("writing report to buf: %w", err)
	}
	// signature is only calculated from 0x0 to 0x2a0
	if err := cert.CheckSignature(x509.ECDSAWithSHA384, buf.Bytes()[:0x2a0], sigEncoded); err != nil {
		return &signatureError{err}
	}

	hasExpectedIDKeyDigest := false
	for _, digest := range v.idKeyDigests {
		if bytes.Equal(digest, report.IDKeyDigest[:]) {
			hasExpectedIDKeyDigest = true
			break
		}
	}

	if !hasExpectedIDKeyDigest {
		switch v.enforceIDKeyDigest {
		case idkeydigest.MAAFallback:
			v.log.Infof("configured idkeydigests %x don't contain reported idkeydigest %x, falling back to MAA validation", v.idKeyDigests, report.IDKeyDigest[:])
			return v.maa.validateToken(ctx, maaToken, v.maaURL, extraData)
		case idkeydigest.WarnOnly:
			v.log.Warnf("configured idkeydigests %x don't contain reported idkeydigest %x", v.idKeyDigests, report.IDKeyDigest[:])
		default:
			return &idKeyError{report.IDKeyDigest[:], v.idKeyDigests}
		}
	}

	return nil
}

// validateVCEKExtensions checks that the certificate extension values in cert match the values described in report.
func validateVCEKExtensions(cert *x509.Certificate, report snpAttestationReport) error {
	var certVersion int
	for _, extension := range cert.Extensions {
		switch extension.Id.String() {
		// check bootloader version
		case "1.3.6.1.4.1.3704.1.3.1":
			{
				_, err := asn1.Unmarshal(extension.Value, &certVersion)
				if err != nil {
					return fmt.Errorf("unmarshalling bootloader version: %w", err)
				}
				if certVersion != int(report.CommittedTCB.Bootloader) {
					return fmt.Errorf("bootloader version %d from report does not match VCEK version %d", int(report.CommittedTCB.Bootloader), certVersion)
				}
			}
		// check TEE version
		case "1.3.6.1.4.1.3704.1.3.2":
			{
				_, err := asn1.Unmarshal(extension.Value, &certVersion)
				if err != nil {
					return fmt.Errorf("unmarshalling tee version: %w", err)
				}
				if certVersion != int(report.CommittedTCB.TEE) {
					return fmt.Errorf("bootloader version %d from report does not match VCEK version %d", int(report.CommittedTCB.TEE), certVersion)
				}
			}
		// check SNP Firmware version
		case "1.3.6.1.4.1.3704.1.3.3":
			{
				_, err := asn1.Unmarshal(extension.Value, &certVersion)
				if err != nil {
					return fmt.Errorf("unmarshalling snp version: %w", err)
				}
				if certVersion != int(report.CommittedTCB.SNP) {
					return fmt.Errorf("bootloader version %d from report does not match VCEK version %d", int(report.CommittedTCB.SNP), certVersion)
				}
			}
		// check microcode version
		case "1.3.6.1.4.1.3704.1.3.8":
			{
				_, err := asn1.Unmarshal(extension.Value, &certVersion)
				if err != nil {
					return fmt.Errorf("unmarshalling microcode version: %w", err)
				}
				if certVersion != int(report.CommittedTCB.Microcode) {
					return fmt.Errorf("bootloader version %d from report does not match VCEK version %d", int(report.CommittedTCB.Microcode), certVersion)
				}
			}
		}
	}

	return nil
}

type azureInstanceInfo struct {
	Vcek              []byte
	CertChain         []byte
	AttestationReport []byte
	RuntimeData       []byte
	MAAToken          string
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

// Reference: https://github.com/AMDESE/sev-guest/blob/main/include/attestation.h
type snpAttestationReport struct {
	Version         uint32       // 0x000
	GuestSVN        uint32       // 0x004
	Policy          guestPolicy  // 0x008
	FamilyID        [16]byte     // 0x010
	ImageID         [16]byte     // 0x020
	VMPL            uint32       // 0x030
	SignatureAlgo   uint32       // 0x034
	CurrentTCB      tcbVersion   // 0x038
	PlatformInfo    uint64       // 0x040
	Flags           uint32       // 0x048
	Reserved0       uint32       // 0x04C
	ReportData      [64]byte     // 0x050
	Measurement     [48]byte     // 0x090
	HostData        [32]byte     // 0x0C0
	IDKeyDigest     [48]byte     // 0x0E0
	AuthorKeyDigest [48]byte     // 0x110
	ReportID        [32]byte     // 0x140
	ReportIDMa      [32]byte     // 0x160
	ReportedTCB     tcbVersion   // 0x180
	_               [24]byte     // 0x188
	ChipID          [64]byte     // 0x1A0
	CommittedTCB    tcbVersion   // 0x1E0
	CurrentBuild    byte         // 0x1E8
	CurrentMinor    byte         // 0x1E9
	CurrentMajor    byte         // 0x1EA
	_               byte         // 0x1EB
	CommittedBuild  byte         // 0x1EC
	CommittedMinor  byte         // 0x1ED
	CommittedMajor  byte         // 0x1EE
	_               byte         // 0x1EF
	LaunchTCB       tcbVersion   // 0x1F0
	_               [168]byte    // 0x1F8
	Signature       snpSignature // 0x2A0
}

type guestPolicy struct {
	AbiMinor       uint8 // 0x0
	AbiMajor       uint8 // 0x8
	ContainerValue byte  // 0x10 - encodes the following four values:
	// Smt          bool // 0x10 - bit 0 in 'ContainerValue'.
	// _            bool // 0x11 - bit 1 in 'ContainerValue'.
	// MigrateMa    bool // 0x12 - bit 2 in 'ContainerValue'.
	// Debug        bool // 0x13 - bit 3 in 'ContainerValue'.
	// SingleSocket bool // 0x14 - bit 4 in 'ContainerValue'.
	_ [5]byte // 0x15
}

func (g *guestPolicy) Debug() bool {
	return (g.ContainerValue & 0b00001000) != 0
}

type tcbVersion struct {
	Bootloader uint8   // 0x0
	TEE        uint8   // 0x10
	_          [4]byte // 0x2F
	SNP        uint8   // 0x37
	Microcode  uint8   // 0x3F
}

func (t *tcbVersion) isVersion(expectedBootloader, expectedTEE, expectedSNP, expectedMicrocode uint8) bool {
	return t.Bootloader >= expectedBootloader && t.TEE >= expectedTEE && t.SNP >= expectedSNP && t.Microcode >= expectedMicrocode
}

func (t *tcbVersion) supersededBy(new tcbVersion) bool {
	return new.Bootloader >= t.Bootloader && new.TEE >= t.TEE && new.SNP >= t.SNP && new.Microcode >= t.Microcode
}

type snpSignature struct {
	R        [72]byte
	S        [72]byte
	Reserved [512 - 144]byte
}
type ecdsaSig struct {
	R, S *big.Int
}

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
	validateToken(context.Context, string, string, []byte) error
}
