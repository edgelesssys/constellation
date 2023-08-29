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
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

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
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
)

// Validator for Azure confidential VM attestation.
type Validator struct {
	variant.AzureSEVSNP
	*vtpm.Validator
	hclValidator hclAkValidator
	maa          maaValidator

	config *config.AzureSEVSNP

	log attestation.Logger
}

// NewValidator initializes a new Azure validator with the provided PCR values.
func NewValidator(cfg *config.AzureSEVSNP, log attestation.Logger) *Validator {
	if log == nil {
		log = nopAttestationLogger{}
	}
	v := &Validator{
		hclValidator: &azureInstanceInfo{},
		maa:          newMAAClient(),
		config:       cfg,
		log:          log,
	}
	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
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

func reverseEndian(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
}

// getTrustedKey establishes trust in the given public key.
// It does so by verifying the SNP attestation document.
func (v *Validator) getTrustedKey(ctx context.Context, attDoc vtpm.AttestationDocument, extraData []byte) (crypto.PublicKey, error) {
	// transform the instanceInfo received from Microsoft into a the verifiable attestation report format.
	var instanceInfo azureInstanceInfo
	if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
		return nil, fmt.Errorf("unmarshalling instanceInfo: %w", err)
	}
	att, err := instanceInfo.attestation()
	if err != nil {
		return nil, fmt.Errorf("parsing attestation report: %w", err)
	}

	// Retrieve the VCEK certificate from the AMD KDS, get the certificate chain for the VCEK
	// certificate, and verify the certificate chain.
	if err := verify.SnpReport(att.Report, &verify.Options{}); err != nil {
		return nil, fmt.Errorf("verifying SNP attestation: %w", err)
	}

	// If the enforcement policy is set to MAAFallback or WarnOnly, the check of the IDKey hashes should not be enforced.
	requireIDBlock := !(v.config.FirmwareSignerConfig.EnforcementPolicy == idkeydigest.MAAFallback ||
		v.config.FirmwareSignerConfig.EnforcementPolicy == idkeydigest.WarnOnly)

	// Checks if the attestation report matches the given constraints.
	// Some constraints are implicitly checked by validate.SnpAttestation:
	// - the report is not expired
	if err := validate.SnpAttestation(att, &validate.Options{
		GuestPolicy: abi.SnpPolicy{
			Debug: false, // Debug means the VM can be decrypted by the host for debugging purposes and thus is not allowed.
		},
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
		// Check if the IDKey hashes in the report are in the list of accepted hashes.
		TrustedIDKeyHashes: v.config.FirmwareSignerConfig.AcceptedKeyDigests,
		// The IDKey hashes should not be checked if the enforcement policy is set to MAAFallback or WarnOnly to prevent
		// an error from being returned because of the TrustedIDKeyHashes validation. In this case, we should perform a
		// custom check of the MAA-specific values later. Right now, this is a double check, since a custom MAA check
		// is performed either way.
		RequireIDBlock: requireIDBlock,
	}); err != nil {
		return nil, fmt.Errorf("validating SNP attestation: %w", err)
	}
	if requireIDBlock {
		// Custom WarnOnly / MAAFallback check of the IDKeyDigests.
		v.checkIDKeyDigests(ctx, att, instanceInfo.MAAToken, extraData)
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

func (v *Validator) checkIDKeyDigests(ctx context.Context, report *spb.Attestation, maaToken string, extraData []byte) error {
	hasExpectedIDKeyDigest := false
	for _, digest := range v.config.FirmwareSignerConfig.AcceptedKeyDigests {
		if bytes.Equal(digest, report.Report.IdKeyDigest) {
			hasExpectedIDKeyDigest = true
			break
		}
	}

	if !hasExpectedIDKeyDigest {
		// IDKeyDigests that were not expected are present, check the enforcement policy and verify against
		// the MAA if necessary.
		switch v.config.FirmwareSignerConfig.EnforcementPolicy {
		case idkeydigest.MAAFallback:
			v.log.Infof(
				"configured idkeydigests %x don't contain reported idkeydigest %x, falling back to MAA validation",
				v.config.FirmwareSignerConfig.AcceptedKeyDigests,
				report.Report.IdKeyDigest,
			)
			return v.maa.validateToken(ctx, v.config.FirmwareSignerConfig.MAAURL, maaToken, extraData)
		case idkeydigest.WarnOnly:
			v.log.Warnf(
				"configured idkeydigests %x don't contain reported idkeydigest %x",
				v.config.FirmwareSignerConfig.AcceptedKeyDigests,
				report.Report.IdKeyDigest,
			)
		default:
			return &idKeyError{report.Report.IdKeyDigest, v.config.FirmwareSignerConfig.AcceptedKeyDigests}
		}
	}

	// No IDKeyDigests that were not expected are present.
	return nil
}

type azureInstanceInfo struct {
	AttestationReport []byte
	RuntimeData       []byte
	MAAToken          string
}

// attestation returns the formatted attestation report.
func (a *azureInstanceInfo) attestation() (*spb.Attestation, error) {
	report, err := abi.ReportToProto(a.AttestationReport)
	if err != nil {
		return nil, fmt.Errorf("converting report to proto: %w", err)
	}

	return &spb.Attestation{
		Report:           report,
	}, nil
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
	validateToken(ctx context.Context, maaURL string, token string, extraData []byte) error
}
