/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/go-tdx-qpl/verification"
	"github.com/edgelesssys/go-tdx-qpl/verification/types"
)

type tdxVerifier interface {
	Verify(ctx context.Context, quote []byte) (types.SGXQuote4, error)
}

// Validator is the TDX attestation validator.
type Validator struct {
	variant.QEMUTDX

	tdx      tdxVerifier
	expected measurements.M

	log attestation.Logger
}

// NewValidator initializes a new TDX Validator.
func NewValidator(cfg *config.QEMUTDX, log attestation.Logger) *Validator {
	if log == nil {
		log = attestation.NOPLogger{}
	}

	return &Validator{
		tdx:      verification.New(),
		expected: cfg.Measurements,
		log:      log,
	}
}

// Validate validates the given attestation document using TDX attestation.
func (v *Validator) Validate(ctx context.Context, attDocRaw []byte, nonce []byte) (userData []byte, err error) {
	v.log.Info("Validating attestation document")
	defer func() {
		if err != nil {
			v.log.Warn(fmt.Sprintf("Failed to validate attestation document: %s", err))
		}
	}()

	var attDoc tdxAttestationDocument
	if err := json.Unmarshal(attDocRaw, &attDoc); err != nil {
		return nil, fmt.Errorf("unmarshaling attestation document: %w", err)
	}

	// Verify the quote.
	quote, err := v.tdx.Verify(ctx, attDoc.RawQuote)
	if err != nil {
		return nil, fmt.Errorf("verifying TDX quote: %w", err)
	}

	// Report data
	extraData := attestation.MakeExtraData(attDoc.UserData, nonce)
	if !attestation.CompareExtraData(quote.Body.ReportData[:], extraData) {
		return nil, fmt.Errorf("report data in TDX quote does not match provided nonce")
	}

	// Convert RTMRs and MRTD to map.
	tdMeasure := make(map[uint32][]byte, 5)
	tdMeasure[0] = quote.Body.MRTD[:]
	for idx := 0; idx < len(quote.Body.RTMR); idx++ {
		tdMeasure[uint32(idx+1)] = quote.Body.RTMR[idx][:]
	}

	// Verify the quote against the expected measurements.
	warnings, errs := v.expected.Compare(tdMeasure)
	for _, warning := range warnings {
		v.log.Warn(warning)
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("measurement validation failed:\n%w", errors.Join(errs...))
	}

	return attDoc.UserData, nil
}
