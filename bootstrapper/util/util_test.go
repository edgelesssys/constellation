package util

import (
	"testing"

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
