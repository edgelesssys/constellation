/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package amdkds

import (
  "fmt"
	"testing"
  "log/slog"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache/amdkds/testdata"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify/trust"
	"github.com/stretchr/testify/assert"
)

func TestCertChain(t *testing.T) {
	testCases := map[string]struct {
		getter  *stubGetter
		wantErr bool
	}{
		"success": {
			getter: &stubGetter{
				log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				ret: testdata.CertChain,
			},
		},
		"getter error": {
			getter: &stubGetter{
				log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				err: assert.AnError,
			},
			wantErr: true,
		},
		"empty cert chain": {
			getter: &stubGetter{
				log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				ret: nil,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Clear the product cert cache to ensure that the test cases are independent.
			trust.ClearProductCertCache()

			assert := assert.New(t)

			kdsClient := NewKDSClient(tc.getter)

			_, _, err := kdsClient.CertChain(abi.NoneReportSigner)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubGetter struct {
	log *slog.Logger
	ret []byte
	err error
}

func (s *stubGetter) Get(url string) ([]byte, error) {
	s.log.Debug(fmt.Sprintf("Request to %s", url))
	return s.ret, s.err
}
