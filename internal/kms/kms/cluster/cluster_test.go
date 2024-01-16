/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cluster

import (
	"context"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/crypto/testvector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestClusterKMS(t *testing.T) {
	testVector := testvector.HKDF0xFF
	assert := assert.New(t)
	require := require.New(t)
	kms, err := New(testVector.Secret, testVector.Salt)
	require.NoError(err)

	keyLower, err := kms.GetDEK(
		context.Background(),
		strings.ToLower(testVector.InfoPrefix+testVector.Info),
		int(testVector.Length),
	)
	assert.NoError(err)
	assert.Equal(testVector.Output, keyLower)

	// output of the KMS should be case sensitive
	keyUpper, err := kms.GetDEK(
		context.Background(),
		strings.ToUpper(testVector.InfoPrefix+testVector.Info),
		int(testVector.Length),
	)
	assert.NoError(err)
	assert.NotEqual(keyLower, keyUpper)
}

func TestVectorsHKDF(t *testing.T) {
	testCases := map[string]struct {
		kek     []byte
		salt    []byte
		dekID   string
		dekSize uint
		wantKey []byte
		wantErr bool
	}{
		"rfc Test Case 1": {
			kek:     testvector.HKDFrfc1.Secret,
			salt:    testvector.HKDFrfc1.Salt,
			dekID:   testvector.HKDFrfc1.Info,
			dekSize: testvector.HKDFrfc1.Length,
			wantKey: testvector.HKDFrfc1.Output,
		},
		"rfc Test Case 2": {
			kek:     testvector.HKDFrfc2.Secret,
			salt:    testvector.HKDFrfc2.Salt,
			dekID:   testvector.HKDFrfc2.Info,
			dekSize: testvector.HKDFrfc2.Length,
			wantKey: testvector.HKDFrfc2.Output,
		},
		"rfc Test Case 3": {
			kek:     testvector.HKDFrfc3.Secret,
			salt:    testvector.HKDFrfc3.Salt,
			dekID:   testvector.HKDFrfc3.Info,
			dekSize: testvector.HKDFrfc3.Length,
			wantKey: testvector.HKDFrfc3.Output,
			wantErr: true,
		},
		"HKDF zero": {
			kek:     testvector.HKDFZero.Secret,
			salt:    testvector.HKDFZero.Salt,
			dekID:   testvector.HKDFZero.InfoPrefix + testvector.HKDFZero.Info,
			dekSize: testvector.HKDFZero.Length,
			wantKey: testvector.HKDFZero.Output,
		},
		"HKDF 0xFF": {
			kek:     testvector.HKDF0xFF.Secret,
			salt:    testvector.HKDF0xFF.Salt,
			dekID:   testvector.HKDF0xFF.InfoPrefix + testvector.HKDF0xFF.Info,
			dekSize: testvector.HKDF0xFF.Length,
			wantKey: testvector.HKDF0xFF.Output,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kms, err := New(tc.kek, tc.salt)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			out, err := kms.GetDEK(context.Background(), tc.dekID, int(tc.dekSize))
			require.NoError(err)
			assert.Equal(tc.wantKey, out)
		})
	}
}
