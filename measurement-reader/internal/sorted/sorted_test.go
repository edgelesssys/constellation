/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sorted

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/stretchr/testify/assert"
)

func TestSortMeasurements(t *testing.T) {
	testCases := map[string]struct {
		measurementType MeasurementType
		input           measurements.M
		want            []Measurement
	}{
		"pre sorted TPM": {
			measurementType: TPM,
			input: measurements.M{
				0: measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength),
				1: measurements.WithAllBytes(0x22, measurements.Enforce, measurements.PCRMeasurementLength),
				2: measurements.WithAllBytes(0x33, measurements.Enforce, measurements.PCRMeasurementLength),
			},
			want: []Measurement{
				{
					Index: "PCR[00]",
					Value: bytes.Repeat([]byte{0x11}, 32),
				},
				{
					Index: "PCR[01]",
					Value: bytes.Repeat([]byte{0x22}, 32),
				},
				{
					Index: "PCR[02]",
					Value: bytes.Repeat([]byte{0x33}, 32),
				},
			},
		},
		"unsorted TPM": {
			measurementType: TPM,
			input: measurements.M{
				1: measurements.WithAllBytes(0x22, measurements.Enforce, measurements.PCRMeasurementLength),
				0: measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength),
				2: measurements.WithAllBytes(0x33, measurements.Enforce, measurements.PCRMeasurementLength),
			},
			want: []Measurement{
				{
					Index: "PCR[00]",
					Value: bytes.Repeat([]byte{0x11}, 32),
				},
				{
					Index: "PCR[01]",
					Value: bytes.Repeat([]byte{0x22}, 32),
				},
				{
					Index: "PCR[02]",
					Value: bytes.Repeat([]byte{0x33}, 32),
				},
			},
		},
		"pre sorted TDX": {
			measurementType: TDX,
			input: measurements.M{
				0: measurements.WithAllBytes(0x11, measurements.Enforce, measurements.TDXMeasurementLength),
				1: measurements.WithAllBytes(0x22, measurements.Enforce, measurements.TDXMeasurementLength),
				2: measurements.WithAllBytes(0x33, measurements.Enforce, measurements.TDXMeasurementLength),
			},
			want: []Measurement{
				{
					Index: "MRTD",
					Value: bytes.Repeat([]byte{0x11}, 48),
				},
				{
					Index: "RTMR[0]",
					Value: bytes.Repeat([]byte{0x22}, 48),
				},
				{
					Index: "RTMR[1]",
					Value: bytes.Repeat([]byte{0x33}, 48),
				},
			},
		},
		"unsorted TDX": {
			measurementType: TDX,
			input: measurements.M{
				1: measurements.WithAllBytes(0x22, measurements.Enforce, measurements.TDXMeasurementLength),
				0: measurements.WithAllBytes(0x11, measurements.Enforce, measurements.TDXMeasurementLength),
				2: measurements.WithAllBytes(0x33, measurements.Enforce, measurements.TDXMeasurementLength),
			},
			want: []Measurement{
				{
					Index: "MRTD",
					Value: bytes.Repeat([]byte{0x11}, 48),
				},
				{
					Index: "RTMR[0]",
					Value: bytes.Repeat([]byte{0x22}, 48),
				},
				{
					Index: "RTMR[1]",
					Value: bytes.Repeat([]byte{0x33}, 48),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got := SortMeasurements(tc.input, tc.measurementType)
			for i := range got {
				assert.Equal(got[i].Index, tc.want[i].Index)
				assert.Equal(got[i].Value, tc.want[i].Value)
			}
		})
	}
}
