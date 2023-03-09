/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted"
	"github.com/stretchr/testify/assert"
)

func TestSortMeasurements(t *testing.T) {
	testCases := map[string]struct {
		input measurements.M
		want  []sorted.Measurement
	}{
		"pre sorted": {
			input: measurements.M{
				0: measurements.WithAllBytes(0x11, false),
				1: measurements.WithAllBytes(0x22, false),
				2: measurements.WithAllBytes(0x33, false),
			},
			want: []sorted.Measurement{
				{
					Index: "MRTD",
					Value: bytes.Repeat([]byte{0x11}, 32),
				},
				{
					Index: "RTMR[0]",
					Value: bytes.Repeat([]byte{0x22}, 32),
				},
				{
					Index: "RTMR[1]",
					Value: bytes.Repeat([]byte{0x33}, 32),
				},
			},
		},
		"unsorted": {
			input: measurements.M{
				1: measurements.WithAllBytes(0x22, false),
				0: measurements.WithAllBytes(0x11, false),
				2: measurements.WithAllBytes(0x33, false),
			},
			want: []sorted.Measurement{
				{
					Index: "MRTD",
					Value: bytes.Repeat([]byte{0x11}, 32),
				},
				{
					Index: "RTMR[0]",
					Value: bytes.Repeat([]byte{0x22}, 32),
				},
				{
					Index: "RTMR[1]",
					Value: bytes.Repeat([]byte{0x33}, 32),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got := sortMeasurements(tc.input)
			for i := range got {
				assert.Equal(got[i].Index, tc.want[i].Index)
				assert.Equal(got[i].Value, tc.want[i].Value)
			}
		})
	}
}
