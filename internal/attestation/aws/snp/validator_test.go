/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTrustedKey(t *testing.T) {
	testCases := map[string]struct {
		akPub   []byte
		info    []byte
		wantErr bool
	}{
		"nul byte docs": {
			akPub:   []byte{0x00, 0x00, 0x00, 0x00},
			info:    []byte{0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		"nil": {
			akPub:   nil,
			info:    nil,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out, err := getTrustedKey(
				context.Background(),
				vtpm.AttestationDocument{
					Attestation: &attest.Attestation{
						AkPub: tc.akPub,
					},
					InstanceInfo: tc.info,
				},
				nil,
			)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			assert.Nil(out)
		})
	}
}

// TestValidateSNPReport has to setup the following to run ValidateSNPReport:
// - parse ARK certificate from constants.go.
// - parse cached ASK certificate from testdata/ask.pem.
// - parse cached SNP report from testdata/report_valid.hex. We only check signature chains so this is self-contained.
// - parse cache VLEK cert from testdata/certs_valid.hex.
func TestValidateSNPReport(t *testing.T) {
	ark, ask, err := loadCachedCertChain()
	require.NoError(t, err)

	testCases := map[string]struct {
		validator  *Validator
		reportPath string
		certsPath  string
		wantErr    bool
	}{
		"success": {
			validator:  &Validator{ark: ark, kdsClient: stubAskGetter{ask}},
			reportPath: "testdata/report_valid.hex",
			certsPath:  "testdata/certs_valid.hex",
		},
		"fail": {
			validator: &Validator{ark: ark, kdsClient: stubAskGetter{ask}},
			// report_invalid = report_valid[0x2A0]+1. This is the start of the signature.
			reportPath: "testdata/report_invalid.hex",
			certsPath:  "testdata/certs_valid.hex",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			info, err := loadInstanceInfo(tc.reportPath, tc.certsPath)
			require.NoError(err)

			err = tc.validator.validateSNPReport(vtpm.AttestationDocument{InstanceInfo: info}, nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

// loadCachedCertChain loads a valid ARK and ASK from the testdata folder.
func loadCachedCertChain() (ark, ask *x509.Certificate, err error) {
	// Replacement is needed as the newline chars are not interpreted due to the backticks. Backticks are required for config formatting.
	tmp := strings.ReplaceAll(constants.AMDRootKey, "\\n", "\n")
	block, _ := pem.Decode([]byte(tmp))
	ark, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	askPEM, err := os.ReadFile("testdata/ask.pem")
	if err != nil {
		return nil, nil, err
	}

	pemBlock, _ := pem.Decode(askPEM)
	ask, err = x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return
}

// loadInstanceInfo loads a valid SNP report and VLEK cert from the testdata folder.
func loadInstanceInfo(reportPath, certsPath string) ([]byte, error) {
	reportEnc, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, err
	}

	reportDec := make([]byte, hex.DecodedLen(len(reportEnc)))
	_, err = hex.Decode(reportDec, reportEnc)
	if err != nil {
		return nil, err
	}

	certsEnc, err := os.ReadFile(certsPath)
	if err != nil {
		return nil, err
	}

	certsDec := make([]byte, hex.DecodedLen(len(certsEnc)))
	_, err = hex.Decode(certsDec, certsEnc)
	if err != nil {
		return nil, err
	}

	info := instanceInfo{Report: reportDec, Certs: certsDec}
	infoRaw, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return infoRaw, nil
}

type stubAskGetter struct {
	ask *x509.Certificate
}

func (s stubAskGetter) GetASK(_ context.Context) (*x509.Certificate, error) {
	return s.ask, nil
}
