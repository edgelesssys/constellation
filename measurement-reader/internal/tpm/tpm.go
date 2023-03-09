/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package tpm reads measurements from a TPM.
package tpm

import (
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

	return sorted.SortMeasurements(m, sorted.TPM), nil
}
