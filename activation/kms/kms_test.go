package kms

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/kmsproto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type stubClient struct {
	getDataKeyErr error
	dataKey       []byte
}

func (c *stubClient) GetDataKey(context.Context, *kmsproto.GetDataKeyRequest, *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error) {
	return &kmsproto.GetDataKeyResponse{DataKey: c.dataKey}, c.getDataKeyErr
}

func TestGetDataKey(t *testing.T) {
	testCases := map[string]struct {
		client  *stubClient
		wantErr bool
	}{
		"GetDataKey success": {
			client: &stubClient{dataKey: []byte{0x1, 0x2, 0x3}},
		},
		"GetDataKey error": {
			client:  &stubClient{getDataKeyErr: errors.New("error")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			listener := bufconn.Listen(1)
			defer listener.Close()

			client := New(
				logger.NewTest(t),
				listener.Addr().String(),
			)

			client.grpc = tc.client

			res, err := client.GetDataKey(context.Background(), "disk-uuid", 32)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.client.dataKey, res)
			}
		})
	}
}
