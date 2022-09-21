/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package trustedlaunch

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedKeyFromSNP(t *testing.T) {
	require := require.New(t)

	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()
	key, err := client.AttestationKeyRSA(tpm)
	require.NoError(err)
	defer key.Close()
	akPub, err := key.PublicArea().Encode()
	require.NoError(err)

	testCases := map[string]struct {
		key          []byte
		instanceInfo []byte
		wantErr      bool
	}{
		"success": {
			key:          akPub,
			instanceInfo: []byte{},
			wantErr:      false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			key, err := trustedKey(tc.key, tc.instanceInfo)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotNil(key)
			}
		})
	}
}

func TestValidateAzureCVM(t *testing.T) {
	testCases := map[string]struct {
		attDoc  vtpm.AttestationDocument
		wantErr bool
	}{
		"success": {
			attDoc:  vtpm.AttestationDocument{},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := validateVM(tc.attDoc)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
