/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/tpm2"
)

// Validator for AWS TPM attestation.
type Validator struct {
	// Embed variant to identify the Validator using varaint.OID().
	variant.AWSSEVSNP
	// Embed validator to implement Validate method for aTLS handshake.
	*vtpm.Validator
	// AMD root key. Root of trust for the ASK used during report validation.
	ark *x509.Certificate
	// kdsClient gets an ASK from somewhere. kdsClient is required for testing.
	kdsClient askGetter
	// reportValidator validates a SNP report. reportValidator is required for testing.
	reportValidator snpReportValidator
}

// NewValidator create a new Validator structure and returns it.
func NewValidator(cfg *config.AWSSEVSNP, log attestation.Logger) *Validator {
	v := &Validator{
		ark:       (*x509.Certificate)(&cfg.AMDRootKey),
		kdsClient: kdsClient{http.DefaultClient},
	}

	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		v.getTrustedKey,
		func(vtpm.AttestationDocument, *attest.MachineState) error { return nil },
		log,
	)
	return v
}

// getTrustedKeys return the public area of the provides attestation key.
// Normally, the key should be verified here, but currently AWS does not provide means to do so.
func (v *Validator) getTrustedKey(ctx context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
	akDigest, err := sha512sum(attDoc.Attestation.GetAkPub())
	if err != nil {
		return nil, fmt.Errorf("calculating hash of attestation key: %w", err)
	}

	if err := v.reportValidator.validate(ctx, attDoc, v.kdsClient, v.ark, akDigest); err != nil {
		return nil, fmt.Errorf("validating SNP report: %w", err)
	}
	// Copied from https://github.com/edgelesssys/constellation/blob/main/internal/attestation/qemu/validator.go
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, err
	}

	return pubArea.Key()
}

// sha512sum PEM-encodes a public key and calculates the SHA512 hash of the encoded key.
func sha512sum(key crypto.PublicKey) ([64]byte, error) {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return [64]byte{}, fmt.Errorf("marshalling public key: %w", err)
	}

	return sha512.Sum512(pub), nil
}

// Validate a given SNP report.
type snpReportValidator interface {
	validate(ctx context.Context, attestation vtpm.AttestationDocument, kdsClient askGetter, ark *x509.Certificate, ak [64]byte) error
}

// Validation logic for the AWS SNP implementation.
type awsValidator struct{}

// validate the report by checking if it has a valid VLEK signature.
// The certificate chain ARK -> ASK -> VLEK is also validated.
// Checks that the report's userData matches the connection's userData.
func (awsValidator) validate(ctx context.Context, attestation vtpm.AttestationDocument, kdsClient askGetter, ark *x509.Certificate, akDigest [64]byte) error {
	var info instanceInfo
	if err := json.Unmarshal(attestation.InstanceInfo, &info); err != nil {
		return fmt.Errorf("unmarshalling instance info: %w", err)
	}

	vlek, err := getVLEK(info.Certs)
	if err != nil {
		return fmt.Errorf("parsing certificates: %w", err)
	}

	ask, err := kdsClient.getASK(ctx)
	if err != nil {
		return fmt.Errorf("getting ASK: %w", err)
	}

	if err := ask.CheckSignatureFrom(ark); err != nil {
		return fmt.Errorf("verifying ASK signature: %w", err)
	}
	if err := vlek.CheckSignatureFrom(ask); err != nil {
		return fmt.Errorf("verifying VLEK signature: %w", err)
	}

	if err := verify.SnpReportSignature(info.Report, vlek); err != nil {
		return fmt.Errorf("verifying snp report signature: %w", err)
	}

	report, err := abi.ReportToProto(info.Report)
	if err != nil {
		return fmt.Errorf("unmarshalling SNP report: %w", err)
	}

	if !bytes.Equal(report.GetReportData(), akDigest[:]) {
		return errors.New("SNP report and attestation statement contain mismatching attestation keys")
	}

	return nil
}

// getVLEK parses the certificate table included in an extended SNP report
// and returns the VLEK certificate.
func getVLEK(certs []byte) (vlek *x509.Certificate, err error) {
	certTable := abi.CertTable{}
	if err = certTable.Unmarshal(certs); err != nil {
		return nil, fmt.Errorf("unmarshalling SNP certificate table: %v", err)
	}

	vlekRaw, err := certTable.GetByGUIDString(abi.VlekGUID)
	if err != nil {
		return nil, fmt.Errorf("getting VLEK certificate: %v", err)
	}

	vlek, err = x509.ParseCertificate(vlekRaw)
	if err != nil {
		return nil, fmt.Errorf("parsing certificate: %w", err)
	}

	return
}

// Query the AMD key distribution service for an AMD signing key.
type kdsClient struct {
	httpClient
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// getASK requests the current certificate chain from the AMD KDS API and returns the ASK.
// There is no information on how to handle CRLs in the official AMD docs.
// Once github.com/google/go-sev-guest adds support to check CRLs for VLEK-based certificate chains
// we can check CRLs here.
func (k kdsClient) getASK(ctx context.Context) (*x509.Certificate, error) {
	// If there are multiple CPU generations (and with that different API paths to call) in the future,
	// we can select the correct path to call based on the information contained in the SNP report.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://kdsintf.amd.com/vlek/v1/Milan/cert_chain", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := k.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting ASK: %w", err)
	}
	defer resp.Body.Close()

	pemChain, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// certificate chain starts with ASK. We hardcode the ARK, so ignore that.
	decodedASK, _ := pem.Decode(pemChain)
	if decodedASK == nil {
		return nil, errors.New("no PEM data found")
	}

	ask, err := x509.ParseCertificate(decodedASK.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing ASK: %w", err)
	}

	return ask, nil
}

// askGetter gets an ASK from somewhere.
type askGetter interface {
	getASK(ctx context.Context) (*x509.Certificate, error)
}
