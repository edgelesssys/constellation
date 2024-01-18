/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tdx

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/tdx/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHCLReport(t *testing.T) {
	testCases := map[string]struct {
		report  []byte
		wantErr bool
	}{
		"success using testdata": {
			report:  testdata.HCLReport,
			wantErr: false,
		},
		"invalid report type": {
			report: func() []byte {
				report := make([]byte, len(testdata.HCLReport))
				copy(report, testdata.HCLReport)
				binary.LittleEndian.PutUint32(report[hclReportTypeOffsetStart:], hclReportTypeInvalid)
				return report
			}(),
			wantErr: true,
		},
		"report too short for HCL report type": {
			report: func() []byte {
				report := make([]byte, hclReportTypeOffsetStart+3)
				copy(report, testdata.HCLReport)
				return report
			}(),
			wantErr: true,
		},
		"report too short for runtime data size": {
			report: func() []byte {
				report := make([]byte, runtimeDataSizeOffset+3)
				copy(report, testdata.HCLReport)
				return report
			}(),
			wantErr: true,
		},
		"runtime data shorter than runtime data size": {
			report: func() []byte {
				report := make([]byte, len(testdata.HCLReport))
				copy(report, testdata.HCLReport)
				// Lets claim the report contains a much larger runtime data entry than it actually does.
				// That way, we can easily test if our code correctly handles reports that are shorter than
				// what they claim to be and avoid panics.
				binary.LittleEndian.PutUint32(report[runtimeDataSizeOffset:], 0xFFFF)
				return report
			}(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			hwReport, runtimeData, err := parseHCLReport(tc.report)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.NotNil(hwReport)
			assert.NotNil(runtimeData)
		})
	}
}

func TestIMDSGetQuote(t *testing.T) {
	testCases := map[string]struct {
		client  *http.Client
		wantErr bool
	}{
		"success": {
			client: newTestClient(func(req *http.Request) *http.Response {
				quote := quoteResponse{
					Quote: "test",
				}
				b, err := json.Marshal(quote)
				require.NoError(t, err)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(b)),
				}
			},
			),
			wantErr: false,
		},
		"bad status code": {
			client: newTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}
			},
			),
			wantErr: true,
		},
		"bad json": {
			client: newTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}
			},
			),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			quoteGetter := imdsQuoteGetter{
				client: tc.client,
			}

			_, err := quoteGetter.getQuote(context.Background(), []byte("test"))
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}
