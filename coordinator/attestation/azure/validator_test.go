package azure

import (
	"testing"

	"github.com/edgelesssys/constellation/coordinator/attestation/simulator"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
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
		errExpected  bool
	}{
		"success": {
			key:          akPub,
			instanceInfo: []byte{},
			errExpected:  false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			key, err := trustedKeyFromSNP(tc.key, tc.instanceInfo)
			if tc.errExpected {
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
		attDoc      vtpm.AttestationDocument
		errExpected bool
	}{
		"success": {
			attDoc:      vtpm.AttestationDocument{},
			errExpected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := validateAzureCVM(tc.attDoc)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
