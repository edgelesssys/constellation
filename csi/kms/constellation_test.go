/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kms

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/kms/kmsproto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type stubKMSClient struct {
	getDataKeyErr error
	dataKey       []byte
}

func (c *stubKMSClient) GetDataKey(context.Context, *kmsproto.GetDataKeyRequest, *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error) {
	return &kmsproto.GetDataKeyResponse{DataKey: c.dataKey}, c.getDataKeyErr
}

func TestConstellationKMS(t *testing.T) {
	testCases := map[string]struct {
		kms     *stubKMSClient
		wantErr bool
	}{
		"GetDataKey success": {
			kms:     &stubKMSClient{dataKey: []byte{0x1, 0x2, 0x3}},
			wantErr: false,
		},
		"GetDataKey error": {
			kms:     &stubKMSClient{getDataKeyErr: errors.New("error")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			listener := bufconn.Listen(1)
			defer listener.Close()

			kms := &ConstellationKMS{
				endpoint: listener.Addr().String(),
				kms:      tc.kms,
			}
			res, err := kms.GetDEK(context.Background(), "data-key", 64)

			if tc.wantErr {
				assert.Error(err)
				assert.Nil(res)
			} else {
				assert.NoError(err)
				assert.NotNil(res)
			}
		})
	}
}
