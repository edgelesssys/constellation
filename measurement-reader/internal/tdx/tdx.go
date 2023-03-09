/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package tdx reads measurements from an Intel TDX guest.
package tdx

import (
	"fmt"
	"sort"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted"
)

// Measurements returns a sorted list of TPM PCR measurements.
func Measurements() ([]sorted.Measurement, error) {
	m, err := tdx.GetSelectedMeasurements(tdx.Open, []int{0, 1, 2, 3, 4})
	if err != nil {
		return nil, err
	}

	return sortMeasurements(m), nil
}

func sortMeasurements(m measurements.M) []sorted.Measurement {
	keys := make([]uint32, 0, len(m))
	for idx := range m {
		keys = append(keys, idx)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var measurements []sorted.Measurement
	for _, idx := range keys {
		expected := m[idx].Expected

		// Index 0   == MRTD
		// Index 1-5 == RTMR[0-4]
		var index string
		if (idx) == 0 {
			index = "MRTD"
		} else {
			index = fmt.Sprintf("RTMR[%01d]", idx-1)
		}

		measurements = append(measurements, sorted.Measurement{
			Index: index,
			Value: expected[:],
		})
	}

	return measurements
}
