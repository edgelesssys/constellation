/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm-tools/proto/tpm"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		// https://github.com/google/go-sev-guest/issues/23
		goleak.IgnoreTopFunction("github.com/golang/glog.(*loggingT).flushDaemon"),
	)
}

func TestExportToFile(t *testing.T) {
	testCases := map[string]struct {
		pcrs    measurements.Measurements
		fs      *afero.Afero
		wantErr bool
	}{
		"file not writeable": {
			pcrs: measurements.Measurements{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			fs:      &afero.Afero{Fs: afero.NewReadOnlyFs(afero.NewMemMapFs())},
			wantErr: true,
		},
		"file writeable": {
			pcrs: measurements.Measurements{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			fs:      &afero.Afero{Fs: afero.NewMemMapFs()},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			path := "test-file"
			err := exportToFile(path, tc.pcrs, tc.fs)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				content, err := tc.fs.ReadFile(path)
				require.NoError(err)

				for _, pcr := range tc.pcrs {
					for _, register := range pcr {
						assert.Contains(string(content), fmt.Sprintf("%#02X", register))
					}
				}
			}
		})
	}
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
								Pcrs: measurements.Measurements{
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
								Pcrs: measurements.Measurements{
									0: measurements.AllBytes(0xAA),
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
				assert.EqualValues(attDoc.Attestation.Quotes[qIdx].Pcrs.Pcrs, pcrs)
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
		pcrs   measurements.Measurements
		format string
	}{
		"json": {
			pcrs: measurements.Measurements{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			format: "json",
		},
		"empty format": {
			pcrs: measurements.Measurements{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			format: "",
		},
		"yaml": {
			pcrs: measurements.Measurements{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			format: "yaml",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var out bytes.Buffer
			err := printPCRs(&out, tc.pcrs, tc.format)
			assert.NoError(err)

			for idx, pcr := range tc.pcrs {
				assert.Contains(out.String(), fmt.Sprintf("%d", idx))
				assert.Contains(out.String(), base64.StdEncoding.EncodeToString(pcr))
			}
		})
	}
}
