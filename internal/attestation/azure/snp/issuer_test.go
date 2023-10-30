/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/go-azguestattestation/maa"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSNPAttestation(t *testing.T) {
	testCases := map[string]struct {
		maaURL    string
		maaToken  string
		apiError  error
		tokenErr  error
		paramsErr error
		wantErr   bool
	}{
		"success without maa": {
			wantErr: false,
		},
		"success with maa": {
			maaURL:   "maaurl",
			maaToken: "maatoken",
			wantErr:  false,
		},
		"api fails": {
			apiError: errors.New(""),
			wantErr:  true,
		},
		"createToken fails": {
			maaURL:   "maaurl",
			tokenErr: errors.New(""),
			wantErr:  true,
		},
		"newParameters fails": {
			paramsErr: errors.New(""),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			imdsClient := stubImdsClient{
				maaURL:   tc.maaURL,
				apiError: tc.apiError,
			}

			params := maa.Parameters{
				SNPReport:   []byte("snpreport"),
				RuntimeData: []byte("runtimedata"),
				VcekCert:    []byte("vcekcert"),
				VcekChain:   []byte("vcekchain"),
			}

			maa := &stubMaaTokenCreator{
				token:     tc.maaToken,
				tokenErr:  tc.tokenErr,
				params:    params,
				paramsErr: tc.paramsErr,
			}

			issuer := Issuer{
				imds: imdsClient,
				maa:  maa,
			}

			data := []byte("data")

			attestationJSON, err := issuer.getInstanceInfo(context.Background(), nil, data)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(data, maa.gotParamsData)
			if tc.maaURL == "" {
				assert.Empty(maa.gotTokenData)
			} else {
				assert.Equal(data, maa.gotTokenData)
			}

			var instanceInfo snp.InstanceInfo
			err = json.Unmarshal(attestationJSON, &instanceInfo)
			require.NoError(err)

			assert.Equal(params.VcekCert, instanceInfo.VCEK)
			assert.Equal(params.VcekChain, instanceInfo.CertChain)
			assert.Equal(params.SNPReport, instanceInfo.AttestationReport)
			assert.Equal(params.RuntimeData, instanceInfo.RuntimeData)
			assert.Equal(tc.maaToken, instanceInfo.MAAToken)
		})
	}
}

// TestGetHCLAttestationKey is a basic smoke test that only checks if getAkPub can be run error free.
// Testing anything else will only verify that the simulator works as expected, since getAkPub
// only retrieves the attestation key from the TPM.
func TestGetHCLAttestationKey(t *testing.T) {
	cgo := os.Getenv("CGO_ENABLED")
	if cgo == "0" {
		t.Skip("skipping test because CGO is disabled and tpm simulator requires it")
	}
	require := require.New(t)
	assert := assert.New(t)

	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()

	// we should receive an error if no key was saved at index `tpmAkIdx`
	_, err = getAttestationKey(tpm)
	assert.Error(err)

	// create a key at the index
	tpmAk, err := tpmclient.NewCachedKey(tpm, tpm2.HandleOwner, tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth | tpm2.FlagNoDA | tpm2.FlagRestricted | tpm2.FlagSign,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}, tpmAkIdx)
	require.NoError(err)
	defer tpmAk.Close()

	// we should now be able to retrieve the key
	_, err = getAttestationKey(tpm)
	assert.NoError(err)
}

type stubImdsClient struct {
	maaURL   string
	apiError error
}

func (c stubImdsClient) getMAAURL(_ context.Context) (string, error) {
	return c.maaURL, c.apiError
}

type stubMaaTokenCreator struct {
	token        string
	tokenErr     error
	gotTokenData []byte

	params        maa.Parameters
	paramsErr     error
	gotParamsData []byte
}

func (s *stubMaaTokenCreator) newParameters(_ context.Context, data []byte, _ io.ReadWriter) (maa.Parameters, error) {
	s.gotParamsData = data
	return s.params, s.paramsErr
}

func (s *stubMaaTokenCreator) createToken(_ context.Context, _ io.ReadWriter, _ string, data []byte, _ maa.Parameters) (string, error) {
	s.gotTokenData = data
	return s.token, s.tokenErr
}
