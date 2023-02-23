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
func GetSelectedMeasurements(open OpenFunc, selection []int) (measurements.M, error) {
	if len(selection) > 5 {
		return nil, fmt.Errorf("invalid measurement selection: max 5 measurements allowed, got %d", len(selection))
	}
	for _, idx := range selection {
		if idx < 0 || idx >= 5 {
			return nil, fmt.Errorf("invalid measurement index %d", idx)
		}
	}

	handle, err := open()
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	tdxMeasurements, err := tdx.ReadMeasurements(handle)
	if err != nil {
		return nil, err
	}

	m := make(measurements.M)
	for _, idx := range selection {
		m[uint32(idx)] = measurements.Measurement{
			Expected: tdxMeasurements[idx][:],
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
