/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package amdkds

import (
	"testing"

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
				log: logger.NewTest(t),
				ret: testdata.CertChain,
			},
		},
		"getter error": {
			getter: &stubGetter{
				log: logger.NewTest(t),
				err: assert.AnError,
			},
			wantErr: true,
		},
		"empty cert chain": {
			getter: &stubGetter{
				log: logger.NewTest(t),
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
	log *logger.Logger
	ret []byte
	err error
}

func (s *stubGetter) Get(url string) ([]byte, error) {
	s.log.Debugf("Request to %s", url)
	return s.ret, s.err
}
