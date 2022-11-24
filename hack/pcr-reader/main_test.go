/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm-tools/proto/tpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestValidatePCRAttDoc(t *testing.T) {
	testCases := map[string]struct {
		attDocRaw []byte
		wantErr   bool
	}{
		"invalid attestation document": {
			attDocRaw: []byte{0x1, 0x2, 0x3},
			wantErr:   true,
		},
		"nil attestation": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{}),
			wantErr:   true,
		},
		"nil quotes": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{
				Attestation: &attest.Attestation{},
			}),
			wantErr: true,
		},
		"invalid PCRs": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					Quotes: []*tpm.Quote{
						{
							Pcrs: &tpm.PCRs{
								Hash: tpm.HashAlgo_SHA256,
								Pcrs: map[uint32][]byte{
									0: {0x1, 0x2, 0x3},
								},
							},
						},
					},
				},
			}),
			wantErr: true,
		},
		"valid PCRs": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					Quotes: []*tpm.Quote{
						{
							Pcrs: &tpm.PCRs{
								Hash: tpm.HashAlgo_SHA256,
								Pcrs: map[uint32][]byte{
									0: bytes.Repeat([]byte{0xAA}, 32),
								},
							},
						},
					},
				},
			}),
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pcrs, err := validatePCRAttDoc(tc.attDocRaw)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)

				attDoc := vtpm.AttestationDocument{}
				require.NoError(json.Unmarshal(tc.attDocRaw, &attDoc))
				qIdx, err := vtpm.GetSHA256QuoteIndex(attDoc.Attestation.Quotes)
				require.NoError(err)

				for pcrIdx, pcrVal := range pcrs {
					assert.Equal(pcrVal.Expected[:], attDoc.Attestation.Quotes[qIdx].Pcrs.Pcrs[pcrIdx])
				}
			}
		})
	}
}

func mustMarshalAttDoc(t *testing.T, attDoc vtpm.AttestationDocument) []byte {
	attDocRaw, err := json.Marshal(attDoc)
	require.NoError(t, err)
	return attDocRaw
}

func TestPrintPCRs(t *testing.T) {
	testCases := map[string]struct {
		format string
	}{
		"json": {
			format: "json",
		},
		"empty format": {
			format: "",
		},
		"yaml": {
			format: "yaml",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			pcrs := measurements.M{
				0: measurements.WithAllBytes(0xAA, true),
				1: measurements.WithAllBytes(0xBB, true),
				2: measurements.WithAllBytes(0xCC, true),
			}

			var out bytes.Buffer
			err := printPCRs(&out, pcrs, tc.format)
			assert.NoError(err)

			for idx, pcr := range pcrs {
				assert.Contains(out.String(), fmt.Sprintf("%d", idx))
				assert.Contains(out.String(), hex.EncodeToString(pcr.Expected[:]))
			}
		})
	}
}
