package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/kmsproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRun(t *testing.T) {
	assert := assert.New(t)
	closeErr := errors.New("closed")

	var err error
	var wg sync.WaitGroup
	server := New(logger.NewTest(t), &stubKMS{})

	creds := atlscredentials.New(nil, nil)

	atlsListener, plainListener := setUpTestListeners()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = server.Run(atlsListener, plainListener, creds)
	}()
	assert.NoError(plainListener.Close())
	wg.Wait()
	assert.Equal(closeErr, err)

	atlsListener, plainListener = setUpTestListeners()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = server.Run(atlsListener, plainListener, creds)
	}()
	assert.NoError(atlsListener.Close())
	wg.Wait()
	assert.Equal(closeErr, err)

	atlsListener, plainListener = setUpTestListeners()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = server.Run(atlsListener, plainListener, creds)
	}()
	go assert.NoError(atlsListener.Close())
	go assert.NoError(plainListener.Close())
	wg.Wait()
	assert.Equal(closeErr, err)
}

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

func setUpTestListeners() (net.Listener, net.Listener) {
	atlsListener := testdialer.NewBufconnDialer().GetListener(net.JoinHostPort("192.0.2.1", "9001"))
	plainListener := testdialer.NewBufconnDialer().GetListener(net.JoinHostPort("192.0.2.1", "9000"))
	return atlsListener, plainListener
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
