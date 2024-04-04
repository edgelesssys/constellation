/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"crypto"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/validate"
	"github.com/google/go-sev-guest/verify"
	"github.com/google/go-sev-guest/verify/trust"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
)

// Validator for GCP SEV-SNP / TPM attestation.
type Validator struct {
	variant.GCPSEVSNP
	*vtpm.Validator
	cfg *config.GCPSEVSNP

	// reportValidator validates a SNP report and is required for testing.
	reportValidator snpReportValidator

	// gceKeyGetter gets the public key of the EK from the GCE metadata API.
	gceKeyGetter func(ctx context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error)

	log attestation.Logger
}

// NewValidator creates a new Validator.
func NewValidator(cfg *config.GCPSEVSNP, log attestation.Logger) (*Validator, error) {
	getGCEKey, err := gcp.TrustedKeyGetter(variant.GCPSEVSNP{}, gcp.NewRESTClient)
	if err != nil {
		return nil, fmt.Errorf("create trusted key getter: %v", err)
	}

	v := &Validator{
		cfg:             cfg,
		reportValidator: &gcpValidator{httpsGetter: trust.DefaultHTTPSGetter(), verifier: &reportVerifierImpl{}, validator: &reportValidatorImpl{}},
		gceKeyGetter:    getGCEKey,
		log:             log,
	}

	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		v.getTrustedKey,
		v.validateCVM,
		log,
	)
	return v, nil
}

// getTrustedKey returns TPM endorsement key provided through the GCE metadata API.
func (v *Validator) getTrustedKey(ctx context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
	ekPub, err := v.gceKeyGetter(ctx, attDoc, nil)
	if err != nil {
		return nil, fmt.Errorf("getting TPM endorsement key: %w", err)
	}

	return ekPub, nil
}

// validateCVM validates the SEV-SNP attestation document.
func (v *Validator) validateCVM(attDoc vtpm.AttestationDocument, state *attest.MachineState) error {
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return fmt.Errorf("decoding public area: %w", err)
	}

	pubKey, err := pubArea.Key()
	if err != nil {
		return fmt.Errorf("getting public key: %w", err)
	}

	akDigest, err := sha512sum(pubKey)
	if err != nil {
		return fmt.Errorf("calculating hash of attestation key: %w", err)
	}

	if err := v.reportValidator.validate(attDoc, (*x509.Certificate)(&v.cfg.AMDSigningKey), (*x509.Certificate)(&v.cfg.AMDRootKey), akDigest, v.cfg, v.log); err != nil {
		return fmt.Errorf("validating SNP report: %w", err)
	}
	return nil
}

// sha512sum PEM-encodes a public key and calculates the SHA512 hash of the encoded key.
func sha512sum(key crypto.PublicKey) ([64]byte, error) {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return [64]byte{}, fmt.Errorf("marshalling public key: %w", err)
	}

	return sha512.Sum512(pub), nil
}

// snpReportValidator validates a given SNP report.
type snpReportValidator interface {
	validate(attestation vtpm.AttestationDocument, ask *x509.Certificate, ark *x509.Certificate, ak [64]byte, config *config.GCPSEVSNP, log attestation.Logger) error
}

// gcpValidator implements the validation for GCP SEV-SNP attestation.
// The properties exist for unittesting.
type gcpValidator struct {
	verifier    reportVerifier
	validator   reportValidator
	httpsGetter trust.HTTPSGetter
}

type reportVerifier interface {
	SnpAttestation(att *sevsnp.Attestation, opts *verify.Options) error
}
type reportValidator interface {
	SnpAttestation(att *sevsnp.Attestation, opts *validate.Options) error
}

type reportValidatorImpl struct{}

func (r *reportValidatorImpl) SnpAttestation(att *sevsnp.Attestation, opts *validate.Options) error {
	return validate.SnpAttestation(att, opts)
}

type reportVerifierImpl struct{}

func (r *reportVerifierImpl) SnpAttestation(att *sevsnp.Attestation, opts *verify.Options) error {
	return verify.SnpAttestation(att, opts)
}

// validate the report by checking if it has a valid VCEK signature.
// The certificate chain ARK -> ASK -> VCEK is also validated.
// Checks that the report's userData matches the connection's userData.
func (a *gcpValidator) validate(attestation vtpm.AttestationDocument, ask *x509.Certificate, ark *x509.Certificate, akDigest [64]byte, config *config.GCPSEVSNP, log attestation.Logger) error {
	var info snp.InstanceInfo
	if err := json.Unmarshal(attestation.InstanceInfo, &info); err != nil {
		return fmt.Errorf("unmarshalling instance info: %w", err)
	}

	certchain := snp.NewCertificateChain(ask, ark)

	att, err := info.AttestationWithCerts(a.httpsGetter, certchain, log)
	if err != nil {
		return fmt.Errorf("getting attestation with certs: %w", err)
	}

	verifyOpts, err := getVerifyOpts(att)
	if err != nil {
		return fmt.Errorf("getting verify options: %w", err)
	}

	if err := a.verifier.SnpAttestation(att, verifyOpts); err != nil {
		return fmt.Errorf("verifying SNP attestation: %w", err)
	}

	validateOpts := &validate.Options{
		// Check that the attestation key's digest is included in the report.
		ReportData: akDigest[:],
		GuestPolicy: abi.SnpPolicy{
			Debug: false, // Debug means the VM can be decrypted by the host for debugging purposes and thus is not allowed.
			SMT:   true,  // Allow Simultaneous Multi-Threading (SMT). Normally, we would want to disable SMT
			// but GCP machines are currently facing issues if it's disabled
		},
		VMPL: new(int), // Checks that Virtual Machine Privilege Level (VMPL) is 0.
		// This checks that the reported LaunchTCB version is equal or greater than the minimum specified in the config.
		// We don't specify Options.MinimumTCB as it only restricts the allowed TCB for Current_ and Reported_TCB.
		// Because we allow Options.ProvisionalFirmware, there is not security gained in also checking Current_ and Reported_TCB.
		// We always have to check Launch_TCB as this value indicated the smallest TCB version a VM has seen during
		// it's lifetime.
		MinimumLaunchTCB: kds.TCBParts{
			BlSpl:    config.BootloaderVersion.Value, // Bootloader
			TeeSpl:   config.TEEVersion.Value,        // TEE (Secure OS)
			SnpSpl:   config.SNPVersion.Value,        // SNP
			UcodeSpl: config.MicrocodeVersion.Value,  // Microcode
		},
		// Check that CurrentTCB >= CommittedTCB.
		PermitProvisionalFirmware: true,
	}

	// Checks if the attestation report matches the given constraints.
	// Some constraints are implicitly checked by validate.SnpAttestation:
	// - the report is not expired
	if err := a.validator.SnpAttestation(att, validateOpts); err != nil {
		return fmt.Errorf("validating SNP attestation: %w", err)
	}

	return nil
}

func getVerifyOpts(att *sevsnp.Attestation) (*verify.Options, error) {
	ask, err := x509.ParseCertificate(att.CertificateChain.AskCert)
	if err != nil {
		return &verify.Options{}, fmt.Errorf("parsing ASK certificate: %w", err)
	}
	ark, err := x509.ParseCertificate(att.CertificateChain.ArkCert)
	if err != nil {
		return &verify.Options{}, fmt.Errorf("parsing ARK certificate: %w", err)
	}

	verifyOpts := &verify.Options{
		DisableCertFetching: true,
		TrustedRoots: map[string][]*trust.AMDRootCerts{
			"Milan": {
				{
					Product: "Milan",
					ProductCerts: &trust.ProductCerts{
						Ask: ask,
						Ark: ark,
					},
				},
			},
		},
	}

	return verifyOpts, nil
}
