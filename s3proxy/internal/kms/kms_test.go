/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kms

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/keyservice/keyserviceproto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type stubClient struct {
	getDataKeyErr error
	dataKey       []byte
}

func (c *stubClient) GetDataKey(context.Context, *keyserviceproto.GetDataKeyRequest, *grpc.ClientConn) (*keyserviceproto.GetDataKeyResponse, error) {
	return &keyserviceproto.GetDataKeyResponse{DataKey: c.dataKey}, c.getDataKeyErr
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
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
