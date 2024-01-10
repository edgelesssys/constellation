/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
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
		hclValidator:         &azure.HCLAkValidator{},
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

// getTrustedKey establishes trust in the given public key.
// It does so by verifying the SNP attestation document.
func (v *Validator) getTrustedKey(ctx context.Context, attDoc vtpm.AttestationDocument, extraData []byte) (crypto.PublicKey, error) {
	trustedAsk := (*x509.Certificate)(&v.config.AMDSigningKey) // ASK, cached by the Join-Service
	trustedArk := (*x509.Certificate)(&v.config.AMDRootKey)    // ARK, specified in the Constellation config

	// fallback certificates, used if not present in THIM response.
	cachedCerts := snp.NewCertificateChain(trustedAsk, trustedArk)

	// transform the instanceInfo received from Microsoft into a verifiable attestation report format.
	var instanceInfo snp.InstanceInfo
	if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
		return nil, fmt.Errorf("unmarshalling instanceInfo: %w", err)
	}

	att, err := instanceInfo.AttestationWithCerts(v.getter, cachedCerts, v.log)
	if err != nil {
		return nil, fmt.Errorf("parsing attestation report: %w", err)
	}

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
						Ark: trustedArk,
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
	if instanceInfo.Azure == nil {
		return nil, errors.New("missing Azure info from instanceInfo")
	}
	if err := v.checkIDKeyDigest(ctx, att, instanceInfo.Azure.MAAToken, extraData); err != nil {
		return nil, fmt.Errorf("checking IDKey digests: %w", err)
	}

	// Decode the public area of the attestation key and validate its trustworthiness.
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, err
	}
	if err = v.hclValidator.Validate(instanceInfo.Azure.RuntimeData, att.Report.ReportData, pubArea.RSAParameters); err != nil {
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
		v.log.Info(
			"Configured idkeydigests %x don't contain reported idkeydigest %x, falling back to MAA validation",
			v.config.FirmwareSignerConfig.AcceptedKeyDigests,
			report.Report.IdKeyDigest,
		)
		return v.maa.validateToken(ctx, v.config.FirmwareSignerConfig.MAAURL, maaToken, extraData)
	case idkeydigest.WarnOnly:
		v.log.Warn(
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

// nopAttestationLogger is a no-op implementation of AttestationLogger.
type nopAttestationLogger struct{}

// Infof is a no-op.
func (nopAttestationLogger) Info(string, ...interface{}) {}

// Warnf is a no-op.
func (nopAttestationLogger) Warn(string, ...interface{}) {}

type maaValidator interface {
	validateToken(ctx context.Context, maaURL string, token string, extraData []byte) error
}

type hclAkValidator interface {
	Validate(runtimeDataRaw []byte, reportData []byte, rsaParameters *tpm2.RSAParams) error
}
