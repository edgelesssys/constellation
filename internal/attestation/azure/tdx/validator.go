/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/google/go-tdx-guest/abi"
	"github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/validate"
	"github.com/google/go-tdx-guest/verify"
	"github.com/google/go-tdx-guest/verify/trust"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
)

// Validator for Azure confidential VM attestation using TDX.
type Validator struct {
	variant.AzureTDX
	*vtpm.Validator
	cfg *config.AzureTDX

	getter       trust.HTTPSGetter
	hclValidator hclAkValidator
}

// NewValidator returns a new Validator for Azure confidential VM attestation using TDX.
func NewValidator(cfg *config.AzureTDX, log attestation.Logger) *Validator {
	v := &Validator{
		cfg:          cfg,
		getter:       trust.DefaultHTTPSGetter(),
		hclValidator: &azure.HCLAkValidator{},
	}

	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		v.getTrustedTPMKey,
		func(vtpm.AttestationDocument, *attest.MachineState) error {
			return nil
		},
		log,
	)

	return v
}

func (v *Validator) getTrustedTPMKey(_ context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
	var instanceInfo instanceInfo
	if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
		return nil, err
	}

	quotePb, err := abi.QuoteToProto(instanceInfo.AttestationReport)
	if err != nil {
		return nil, err
	}
	quote, ok := quotePb.(*tdx.QuoteV4)
	if !ok {
		return nil, fmt.Errorf("unexpected quote type: %T", quote)
	}

	if err := v.validateQuote(quote); err != nil {
		return nil, err
	}

	// Decode the public area of the attestation key and validate its trustworthiness.
	pubArea, err := tpm2.DecodePublic(attDoc.Attestation.AkPub)
	if err != nil {
		return nil, err
	}
	if err = v.hclValidator.Validate(instanceInfo.RuntimeData, quote.TdQuoteBody.ReportData, pubArea.RSAParameters); err != nil {
		return nil, fmt.Errorf("validating HCLAkPub: %w", err)
	}

	return pubArea.Key()
}

func (v *Validator) validateQuote(tdxQuote *tdx.QuoteV4) error {
	roots := x509.NewCertPool()
	roots.AddCert((*x509.Certificate)(&v.cfg.IntelRootKey))

	if err := verify.TdxQuote(tdxQuote, &verify.Options{
		CheckRevocations: true,
		GetCollateral:    true,
		TrustedRoots:     roots,
		Getter:           v.getter,
	}); err != nil {
		return err
	}

	if err := validate.TdxQuote(tdxQuote, &validate.Options{
		HeaderOptions: validate.HeaderOptions{
			MinimumQeSvn:  v.cfg.QESVN,
			MinimumPceSvn: v.cfg.PCESVN,
			QeVendorID:    v.cfg.QEVendorID,
		},
		TdQuoteBodyOptions: validate.TdQuoteBodyOptions{
			MinimumTeeTcbSvn: v.cfg.TEETCBSVN,
			MrSeam:           v.cfg.MRSeam,
			Xfam:             v.cfg.XFAM,
		},
	}); err != nil {
		return err
	}

	return nil
}

type hclAkValidator interface {
	Validate(runtimeDataRaw []byte, reportData []byte, rsaParameters *tpm2.RSAParams) error
}
