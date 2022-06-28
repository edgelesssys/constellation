package kmsapi

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	log := logger.NewTest(t)

	kms := &stubKMS{derivedKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5}}
	api := New(log, kms)

	res, err := api.GetDataKey(context.Background(), &kmsproto.GetDataKeyRequest{DataKeyId: "1", Length: 32})
	require.NoError(err)
	assert.Equal(kms.derivedKey, res.DataKey)

	// Test no data key id
	res, err = api.GetDataKey(context.Background(), &kmsproto.GetDataKeyRequest{Length: 32})
	require.Error(err)
	assert.Nil(res)

	// Test no / zero key length
	res, err = api.GetDataKey(context.Background(), &kmsproto.GetDataKeyRequest{DataKeyId: "1"})
	require.Error(err)
	assert.Nil(res)

	// Test derive key error
	api = New(log, &stubKMS{deriveKeyErr: errors.New("error")})
	res, err = api.GetDataKey(context.Background(), &kmsproto.GetDataKeyRequest{DataKeyId: "1", Length: 32})
	assert.Error(err)
	assert.Nil(res)
}

type stubKMS struct {
	masterKey    []byte
	derivedKey   []byte
	deriveKeyErr error
}

func (c *stubKMS) CreateKEK(ctx context.Context, keyID string, kek []byte) error {
	c.masterKey = kek
	return nil
}

func (c *stubKMS) GetDEK(ctx context.Context, kekID string, dekID string, dekSize int) ([]byte, error) {
	if c.deriveKeyErr != nil {
		return nil, c.deriveKeyErr
	}
	return c.derivedKey, nil
}
