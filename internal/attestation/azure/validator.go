package azure

import (
	"bytes"
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

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	internalCrypto "github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/google/go-tpm/tpm2"
)

// AMD root key. Received from the AMD Key Distribution System API (KDS).
const arkPEM = "-----BEGIN CERTIFICATE-----\nMIIGYzCCBBKgAwIBAgIDAQAAMEYGCSqGSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAIC\nBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZIAWUDBAICBQCiAwIBMKMDAgEBMHsxFDAS\nBgNVBAsMC0VuZ2luZWVyaW5nMQswCQYDVQQGEwJVUzEUMBIGA1UEBwwLU2FudGEg\nQ2xhcmExCzAJBgNVBAgMAkNBMR8wHQYDVQQKDBZBZHZhbmNlZCBNaWNybyBEZXZp\nY2VzMRIwEAYDVQQDDAlBUkstTWlsYW4wHhcNMjAxMDIyMTcyMzA1WhcNNDUxMDIy\nMTcyMzA1WjB7MRQwEgYDVQQLDAtFbmdpbmVlcmluZzELMAkGA1UEBhMCVVMxFDAS\nBgNVBAcMC1NhbnRhIENsYXJhMQswCQYDVQQIDAJDQTEfMB0GA1UECgwWQWR2YW5j\nZWQgTWljcm8gRGV2aWNlczESMBAGA1UEAwwJQVJLLU1pbGFuMIICIjANBgkqhkiG\n9w0BAQEFAAOCAg8AMIICCgKCAgEA0Ld52RJOdeiJlqK2JdsVmD7FktuotWwX1fNg\nW41XY9Xz1HEhSUmhLz9Cu9DHRlvgJSNxbeYYsnJfvyjx1MfU0V5tkKiU1EesNFta\n1kTA0szNisdYc9isqk7mXT5+KfGRbfc4V/9zRIcE8jlHN61S1ju8X93+6dxDUrG2\nSzxqJ4BhqyYmUDruPXJSX4vUc01P7j98MpqOS95rORdGHeI52Naz5m2B+O+vjsC0\n60d37jY9LFeuOP4Meri8qgfi2S5kKqg/aF6aPtuAZQVR7u3KFYXP59XmJgtcog05\ngmI0T/OitLhuzVvpZcLph0odh/1IPXqx3+MnjD97A7fXpqGd/y8KxX7jksTEzAOg\nbKAeam3lm+3yKIcTYMlsRMXPcjNbIvmsBykD//xSniusuHBkgnlENEWx1UcbQQrs\n+gVDkuVPhsnzIRNgYvM48Y+7LGiJYnrmE8xcrexekBxrva2V9TJQqnN3Q53kt5vi\nQi3+gCfmkwC0F0tirIZbLkXPrPwzZ0M9eNxhIySb2npJfgnqz55I0u33wh4r0ZNQ\neTGfw03MBUtyuzGesGkcw+loqMaq1qR4tjGbPYxCvpCq7+OgpCCoMNit2uLo9M18\nfHz10lOMT8nWAUvRZFzteXCm+7PHdYPlmQwUw3LvenJ/ILXoQPHfbkH0CyPfhl1j\nWhJFZasCAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMB0GA1UdDgQWBBSFrBrRQ/fI\nrFXUxR1BSKvVeErUUzAPBgNVHRMBAf8EBTADAQH/MDoGA1UdHwQzMDEwL6AtoCuG\nKWh0dHBzOi8va2RzaW50Zi5hbWQuY29tL3ZjZWsvdjEvTWlsYW4vY3JsMEYGCSqG\nSIb3DQEBCjA5oA8wDQYJYIZIAWUDBAICBQChHDAaBgkqhkiG9w0BAQgwDQYJYIZI\nAWUDBAICBQCiAwIBMKMDAgEBA4ICAQC6m0kDp6zv4Ojfgy+zleehsx6ol0ocgVel\nETobpx+EuCsqVFRPK1jZ1sp/lyd9+0fQ0r66n7kagRk4Ca39g66WGTJMeJdqYriw\nSTjjDCKVPSesWXYPVAyDhmP5n2v+BYipZWhpvqpaiO+EGK5IBP+578QeW/sSokrK\ndHaLAxG2LhZxj9aF73fqC7OAJZ5aPonw4RE299FVarh1Tx2eT3wSgkDgutCTB1Yq\nzT5DuwvAe+co2CIVIzMDamYuSFjPN0BCgojl7V+bTou7dMsqIu/TW/rPCX9/EUcp\nKGKqPQ3P+N9r1hjEFY1plBg93t53OOo49GNI+V1zvXPLI6xIFVsh+mto2RtgEX/e\npmMKTNN6psW88qg7c1hTWtN6MbRuQ0vm+O+/2tKBF2h8THb94OvvHHoFDpbCELlq\nHnIYhxy0YKXGyaW1NjfULxrrmxVW4wcn5E8GddmvNa6yYm8scJagEi13mhGu4Jqh\n3QU3sf8iUSUr09xQDwHtOQUVIqx4maBZPBtSMf+qUDtjXSSq8lfWcd8bLr9mdsUn\nJZJ0+tuPMKmBnSH860llKk+VpVQsgqbzDIvOLvD6W1Umq25boxCYJ+TuBoa4s+HH\nCViAvgT9kf/rBq1d+ivj6skkHxuzcxbk1xv6ZGxrteJxVH7KlX7YRdZ6eARKwLe4\nAFZEAwoKCQ==\n-----END CERTIFICATE-----\n"

// Validator for Azure confidential VM attestation.
type Validator struct {
	oid.Azure
	*vtpm.Validator
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(pcrs map[uint32][]byte, enforcedPCRs []uint32) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			enforcedPCRs,
			trustedKeyFromSNP(&azureInstanceInfo{}),
			validateAzureCVM,
			vtpm.VerifyPKCS1v15,
		),
	}
}

type signatureError struct {
	innerError error
}

func (e *signatureError) Unwrap() error {
	return e.innerError
}

func (e *signatureError) Error() string {
	return fmt.Sprintf("signature validation failed: %v", e.innerError)
}

type askError struct {
	innerError error
}

func (e *askError) Unwrap() error {
	return e.innerError
}

func (e *askError) Error() string {
	return fmt.Sprintf("validating ASK: %v", e.innerError)
}

type vcekError struct {
	innerError error
}

func (e *vcekError) Unwrap() error {
	return e.innerError
}

func (e *vcekError) Error() string {
	return fmt.Sprintf("validating VCEK: %v", e.innerError)
}

// trustedKeyFromSNP establishes trust in the given public key.
// It does so by verifying the SNP attestation statement in instanceInfo.
func trustedKeyFromSNP(hclAk HCLAkValidator) func(akPub, instanceInfoRaw []byte) (crypto.PublicKey, error) {
	return func(akPub, instanceInfoRaw []byte) (crypto.PublicKey, error) {
		var instanceInfo azureInstanceInfo
		if err := json.Unmarshal(instanceInfoRaw, &instanceInfo); err != nil {
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

		if err = validateSNPReport(vcek, report); err != nil {
			return nil, fmt.Errorf("validating SNP report: %w", err)
		}

		pubArea, err := tpm2.DecodePublic(akPub)
		if err != nil {
			return nil, err
		}

		if err = hclAk.validateAk(instanceInfo.RuntimeData, report.ReportData[:], pubArea.RSAParameters); err != nil {
			return nil, fmt.Errorf("validating HCLAkPub: %w", err)
		}

		return pubArea.Key()
	}
}

func reverseEndian(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
}

// validateAzureCVM is a stub, since SEV-SNP attestation is already verified in trustedKeyFromSNP().
func validateAzureCVM(attestation vtpm.AttestationDocument) error {
	return nil
}

func newSNPReportFromBytes(reportRaw []byte) (snpAttestationReport, error) {
	var report snpAttestationReport
	if err := binary.Read(bytes.NewReader(reportRaw), binary.LittleEndian, &report); err != nil {
		return snpAttestationReport{}, fmt.Errorf("reading attestation report: %w", err)
	}

	return report, nil
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

func validateSNPReport(cert *x509.Certificate, report snpAttestationReport) error {
	sig_r := report.Signature.R[:]
	sig_s := report.Signature.S[:]

	// Table 107 in https://www.amd.com/system/files/TechDocs/56860.pdf mentions little endian signature components.
	// They come out of the certificate as big endian.
	reverseEndian(sig_r)
	reverseEndian(sig_s)

	r := new(big.Int).SetBytes(sig_r)
	s := new(big.Int).SetBytes(sig_s)
	sequence := ecdsaSig{r, s}
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

	return nil
}

type azureInstanceInfo struct {
	Vcek              []byte
	CertChain         []byte
	AttestationReport []byte
	RuntimeData       []byte
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
	if !bytes.Equal(sum[:], reportData[:len(sum)]) {
		return errors.New("unexpected runtimeData digest in TPM")
	}

	if len(runtimeData.Keys) < 1 {
		return errors.New("did not receive any keys in runtime data")
	}
	rawN, err := base64.RawURLEncoding.DecodeString(runtimeData.Keys[0].N)
	if err != nil {
		return err
	}
	if !bytes.Equal(rawN, rsaParameters.ModulusRaw) {
		return fmt.Errorf("unexpected modulus value in TPM")
	}

	rawE, err := base64.RawURLEncoding.DecodeString(runtimeData.Keys[0].E)
	if err != nil {
		return err
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

type HCLAkValidator interface {
	validateAk(runtimeDataRaw []byte, reportData []byte, rsaParameters *tpm2.RSAParams) error
}

type snpSignature struct {
	R        [72]byte
	S        [72]byte
	Reserved [512 - 144]byte
}

// Reference: https://github.com/AMDESE/sev-guest/blob/main/include/attestation.h
type snpAttestationReport struct {
	Version         uint32       /* 0x000 */
	GuestSvn        uint32       /* 0x004 */
	Policy          uint64       /* 0x008 */
	FamilyId        [16]byte     /* 0x010 */
	ImageId         [16]byte     /* 0x020 */
	Vmpl            uint32       /* 0x030 */
	SignatureAlgo   uint32       /* 0x034 */
	PlatformVersion uint64       /* 0x038 */
	PlatformInfo    uint64       /* 0x040 */
	Flags           uint32       /* 0x048 */
	Reserved0       uint32       /* 0x04C */
	ReportData      [64]byte     /* 0x050 */
	Measurement     [48]byte     /* 0x090 */
	HostData        [32]byte     /* 0x0C0 */
	IdKeyDigest     [48]byte     /* 0x0E0 */
	AuthorKeyDigest [48]byte     /* 0x110 */
	ReportId        [32]byte     /* 0x140 */
	ReportIdMa      [32]byte     /* 0x160 */
	ReportedTcb     uint64       /* 0x180 */
	Reserved1       [24]byte     /* 0x188 */
	ChipId          [64]byte     /* 0x1A0 */
	Reserved2       [192]byte    /* 0x1E0 */
	Signature       snpSignature /* 0x2A0 */
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
