/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/google/go-tpm-tools/client"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateAk tests the attestation key validation with a simulated TPM device.
func TestValidateAk(t *testing.T) {
	cgo := os.Getenv("CGO_ENABLED")
	if cgo == "0" {
		t.Skip("skipping test because CGO is disabled and tpm simulator requires it")
	}

	int32ToBytes := func(val uint32) []byte {
		r := make([]byte, 4)
		binary.PutUvarint(r, uint64(val))
		return r
	}

	require := require.New(t)

	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()
	key, err := client.AttestationKeyRSA(tpm)
	require.NoError(err)
	defer key.Close()

	e := base64.RawURLEncoding.EncodeToString(int32ToBytes(key.PublicArea().RSAParameters.ExponentRaw))
	n := base64.RawURLEncoding.EncodeToString(key.PublicArea().RSAParameters.ModulusRaw)
	ak := akPub{E: e, N: n}
	rtData := runtimeData{PublicPart: []akPub{ak}}

	defaultRuntimeDataRaw, err := json.Marshal(rtData)
	require.NoError(err)
	defaultInstanceInfo := snp.InstanceInfo{Azure: &snp.AzureInstanceInfo{RuntimeData: defaultRuntimeDataRaw}}

	sig := sha256.Sum256(defaultRuntimeDataRaw)
	defaultReportData := sig[:]
	defaultRsaParams := key.PublicArea().RSAParameters

	testCases := map[string]struct {
		instanceInfo   snp.InstanceInfo
		runtimeDataRaw []byte
		reportData     []byte
		rsaParameters  *tpm2.RSAParams
		wantErr        bool
	}{
		"success": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     defaultReportData,
			rsaParameters:  defaultRsaParams,
		},
		"invalid json": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: []byte(""),
			reportData:     defaultReportData,
			rsaParameters:  defaultRsaParams,
			wantErr:        true,
		},
		"invalid hash": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     bytes.Repeat([]byte{0}, 64),
			rsaParameters:  defaultRsaParams,
			wantErr:        true,
		},
		"invalid E": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     defaultReportData,
			rsaParameters: func() *tpm2.RSAParams {
				tmp := *defaultRsaParams
				tmp.ExponentRaw = 1
				return &tmp
			}(),
			wantErr: true,
		},
		"invalid N": {
			instanceInfo:   defaultInstanceInfo,
			runtimeDataRaw: defaultRuntimeDataRaw,
			reportData:     defaultReportData,
			rsaParameters: func() *tpm2.RSAParams {
				tmp := *defaultRsaParams
				tmp.ModulusRaw = []byte{0, 1, 2, 3}
				return &tmp
			}(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ak := HCLAkValidator{}
			err = ak.Validate(tc.runtimeDataRaw, tc.reportData, tc.rsaParameters)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

// TestGetHCLAttestationKey is a basic smoke test that only checks if GetAttestationKey can be run error free.
// Testing anything else will only verify that the simulator works as expected, since GetAttestationKey
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
	_, err = GetAttestationKey(tpm)
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
	_, err = GetAttestationKey(tpm)
	assert.NoError(err)
}
