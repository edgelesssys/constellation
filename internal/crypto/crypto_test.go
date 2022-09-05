/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package crypto

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/crypto/testvector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDeriveKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	key, err := DeriveKey([]byte("secret"), []byte("salt"), nil, 32)
	assert.NoError(err)
	assert.Len(key, 32)

	key1, err := DeriveKey([]byte("secret"), []byte("salt"), []byte("first"), 32)
	require.NoError(err)
	key2, err := DeriveKey([]byte("secret"), []byte("salt"), []byte("first"), 32)
	require.NoError(err)
	assert.Equal(key1, key2)

	key3, err := DeriveKey([]byte("secret"), []byte("salt"), []byte("second"), 32)
	require.NoError(err)
	assert.NotEqual(key1, key3)

	zeroInput := testvector.HKDFZero
	out, err := DeriveKey(zeroInput.Secret, zeroInput.Salt, []byte(zeroInput.InfoPrefix+zeroInput.Info), zeroInput.Length)
	require.NoError(err)
	assert.Equal(zeroInput.Output, out)

	fInput := testvector.HKDF0xFF
	out, err = DeriveKey(fInput.Secret, fInput.Salt, []byte(fInput.InfoPrefix+fInput.Info), fInput.Length)
	require.NoError(err)
	assert.Equal(fInput.Output, out)
}

func TestVectorsHKDF(t *testing.T) {
	testCases := map[string]struct {
		secret  []byte
		salt    []byte
		info    []byte
		length  uint
		wantKey []byte
	}{
		"rfc Test Case 1": {
			secret:  testvector.HKDFrfc1.Secret,
			salt:    testvector.HKDFrfc1.Salt,
			info:    []byte(testvector.HKDFrfc1.Info),
			length:  testvector.HKDFrfc1.Length,
			wantKey: testvector.HKDFrfc1.Output,
		},
		"rfc Test Case 2": {
			secret:  testvector.HKDFrfc2.Secret,
			salt:    testvector.HKDFrfc2.Salt,
			info:    []byte(testvector.HKDFrfc2.Info),
			length:  testvector.HKDFrfc2.Length,
			wantKey: testvector.HKDFrfc2.Output,
		},
		"rfc Test Case 3": {
			secret:  testvector.HKDFrfc3.Secret,
			salt:    testvector.HKDFrfc3.Salt,
			info:    []byte(testvector.HKDFrfc3.Info),
			length:  testvector.HKDFrfc3.Length,
			wantKey: testvector.HKDFrfc3.Output,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			out, err := DeriveKey(tc.secret, tc.salt, tc.info, tc.length)
			require.NoError(err)
			assert.Equal(tc.wantKey, out)
		})
	}
}

func TestGenerateCertificateSerialNumber(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s1, err := GenerateCertificateSerialNumber()
	require.NoError(err)
	s2, err := GenerateCertificateSerialNumber()
	require.NoError(err)
	assert.NotEqual(s1, s2)
}

func TestGenerateRandomBytes(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	n1, err := GenerateRandomBytes(32)
	require.NoError(err)
	assert.Len(n1, 32)

	n2, err := GenerateRandomBytes(32)
	require.NoError(err)

	assert.Equal(len(n1), len(n2))
	assert.NotEqual(n1, n2)

	n3, err := GenerateRandomBytes(16)
	require.NoError(err)
	assert.Len(n3, 16)
}
