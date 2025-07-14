/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package tdx reads measurements from an Intel TDX guest.
package tdx

import (
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted"
)

// Measurements returns a sorted list of TDX runtime measurements.
func Measurements() ([]sorted.Measurement, error) {
	m, err := tdx.GetSelectedMeasurements(tdx.Open, []int{0, 1, 2, 3, 4})
	if err != nil {
		return nil, err
	}

	return sorted.SortMeasurements(m, sorted.TDX), nil
}
