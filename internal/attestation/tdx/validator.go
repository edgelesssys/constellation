/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/edgelesssys/go-tdx-qpl/verification"
)

// Validator is the TDX attestation validator.
type Validator struct {
	oid.QEMUTDX

	tdx      *verification.TDXVerifier
	expected measurements.M

	log vtpm.AttestationLogger
}

// NewValidator initializes a new TDX Validator.
func NewValidator(measurements measurements.M, log vtpm.AttestationLogger) *Validator {
	return &Validator{
		tdx: verification.New(),
	}
}

// Validate validates the given attestation document using TDX attestation.
func (v *Validator) Validate(attDocRaw []byte, nonce []byte) (userData []byte, err error) {
	v.log.Infof("Validating attestation document")
	defer func() {
		if err != nil {
			v.log.Warnf("Failed to validate attestation document: %s", err)
		}
	}()

	var attDoc tdxAttestationDocument
	if err := json.Unmarshal(attDocRaw, &attDoc); err != nil {
		return nil, fmt.Errorf("unmarshaling attestation document: %w", err)
	}

	// Verify the quote.
	quote, err := v.tdx.Verify(context.Background(), attDoc.RawQuote)
	if err != nil {
		return nil, fmt.Errorf("verifying TDX quote: %w", err)
	}

	// TODO: switch to using makeExtraData after rebasing
	// if !bytes.Equal(quote.Body.ReportData[:], makeExtraData(attDoc.UserData, nonce)) {
	if !bytes.Equal(quote.Body.ReportData[:], nonce) {
		return nil, fmt.Errorf("report data in TDX quote does not match provided nonce")
	}

	// Convert RTMRs to map.
	rtmrs := make(map[uint32][]byte)
	for idx, rtmr := range quote.Body.RTMR {
		rtmrs[uint32(idx)] = rtmr[:]
	}

	// Verify the quote against the expected measurements.
	for idx, ex := range v.expected {
		if !bytes.Equal(ex.Expected, rtmrs[idx]) {
			if !ex.WarnOnly {
				return nil, fmt.Errorf("untrusted RTMR value at index %d", idx)
			}
			v.log.Warnf("Encountered untrusted RTMR value at index %d", idx)
		}
	}

	return attDoc.UserData, nil
}
