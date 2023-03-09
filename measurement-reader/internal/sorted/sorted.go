/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package sorted defines a type for print-friendly sorted measurements and allows sorting TPM and TDX measurements.
package sorted

import (
	"fmt"
	"sort"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
)

// Measurement wraps a measurement custom index and value.
type Measurement struct {
	Index string
	Value []byte
}

// MeasurementType are the supported attestation types we can sort.
type MeasurementType uint32

const (
	TPM MeasurementType = iota
	TDX
)

// SortMeasurements returns the sorted measurements for either TPM or TDX measurements.
func SortMeasurements(m measurements.M, measurementType MeasurementType) []Measurement {
	if measurementType != TPM && measurementType != TDX {
		return nil
	}

	keys := make([]uint32, 0, len(m))
	for idx := range m {
		keys = append(keys, idx)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var sortedMeasurements []Measurement

	for _, idx := range keys {
		var index string
		switch measurementType {
		case TPM:
			index = fmt.Sprintf("PCR[%02d]", idx)
		case TDX:
			// idx 0 is MRTD
			if idx == 0 {
				index = "MRTD"
				break
			}
			// RTMR 0 starts at idx 1, so we have to subtract by one here.
			index = fmt.Sprintf("RTMR[%01d]", idx-1)
		}

		expected := m[idx].Expected
		sortedMeasurements = append(sortedMeasurements, Measurement{
			Index: index,
			Value: expected[:],
		})
	}

	return sortedMeasurements
}
