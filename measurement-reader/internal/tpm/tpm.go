/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package tpm reads measurements from a TPM.
package tpm

import (
	"fmt"
	"sort"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted"
	tpmClient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
)

// Measurements returns a sorted list of TPM PCR measurements.
func Measurements() ([]sorted.Measurement, error) {
	m, err := vtpm.GetSelectedMeasurements(vtpm.OpenVTPM, tpmClient.FullPcrSel(tpm2.AlgSHA256))
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
		measurements = append(measurements, sorted.Measurement{
			Index: fmt.Sprintf("PCR[%02d]", idx),
			Value: expected[:],
		})
	}

	return measurements
}
