/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package TDX implements attestation for Intel TDX.
package tdx

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/go-tdx-qpl/tdx"
)

type tdxAttestationDocument struct {
	// RawQuote is the raw TDX quote.
	RawQuote []byte
	// UserData is the user data that was passed to the enclave and was included in the quote.
	UserData []byte
}

// OpenFunc is a function that opens the TDX device.
type OpenFunc func() (tdx.Device, error)

// GetSelectedMeasurements returns the selected measurements from the RTMRs.
func GetSelectedMeasurements(open OpenFunc, rtmrSelection []int) (measurements.M, error) {
	if len(rtmrSelection) > 4 {
		return nil, fmt.Errorf("invalid RTMR selection: max 4 RTMRs allowed, got %d", len(rtmrSelection))
	}

	handle, err := open()
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	rtmr, err := tdx.ReadRTMRs(handle)
	if err != nil {
		return nil, err
	}

	m := make(measurements.M)
	for _, rtmrIdx := range rtmrSelection {
		if rtmrIdx < 0 || rtmrIdx >= len(rtmr) {
			return nil, fmt.Errorf("invalid RTMR index %d", rtmrIdx)
		}
		m[uint32(rtmrIdx)] = measurements.Measurement{
			Expected: rtmr[rtmrIdx][:],
		}
	}

	return m, nil
}

// Available returns true if the TDX device is available and can be opened.
func Available() bool {
	handle, err := tdx.Open()
	if err != nil {
		return false
	}
	defer handle.Close()
	return true
}
