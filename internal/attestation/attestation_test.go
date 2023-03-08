/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestation

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/crypto/testvector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveClusterID(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	testvector := testvector.HKDFClusterID
	clusterID, err := DeriveClusterID(testvector.Secret, testvector.Salt)
	require.NoError(err)
	assert.Equal(testvector.Output, clusterID)

	clusterIDdiff, err := DeriveClusterID(testvector.Secret, []byte("different-salt"))
	require.NoError(err)
	assert.NotEqual(clusterID, clusterIDdiff)

	clusterIDdiff, err = DeriveClusterID([]byte("different-secret"), testvector.Salt)
	require.NoError(err)
	assert.NotEqual(clusterID, clusterIDdiff)
}

func TestCompareExtraData(t *testing.T) {
	testCases := map[string]struct {
		ExtraData1 []byte
		ExtraData2 []byte
		Expected   bool
	}{
		"equal": {
			ExtraData1: bytes.Repeat([]byte{0xAB}, 32),
			ExtraData2: bytes.Repeat([]byte{0xAB}, 32),
			Expected:   true,
		},
		"unequal": {
			ExtraData1: bytes.Repeat([]byte{0xAB}, 32),
			ExtraData2: bytes.Repeat([]byte{0xCD}, 32),
			Expected:   false,
		},
		"unequal length": {
			ExtraData1: bytes.Repeat([]byte{0xAB}, 32),
			ExtraData2: bytes.Repeat([]byte{0xAB}, 64),
			Expected:   false,
		},
		"unequal length, padded with 0": {
			ExtraData1: []byte{0xAB, 0xAB, 0xAB, 0xAB},
			ExtraData2: []byte{0xAB, 0xAB, 0xAB, 0xAB, 0x00, 0x00, 0x00, 0x00},
			Expected:   true,
		},
		"unequal length, prefixed with 0": {
			ExtraData1: []byte{0x00, 0x00, 0x00, 0x00, 0xAB, 0xAB, 0xAB, 0xAB},
			ExtraData2: []byte{0xAB, 0xAB, 0xAB, 0xAB},
			Expected:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			actual := CompareExtraData(tc.ExtraData1, tc.ExtraData2)
			assert.Equal(tc.Expected, actual)
			actual = CompareExtraData(tc.ExtraData2, tc.ExtraData1)
			assert.Equal(tc.Expected, actual)
		})
	}
}
