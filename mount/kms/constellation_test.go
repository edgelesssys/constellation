package kms

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type stubVPNClient struct {
	getDataKeyErr error
	dataKey       []byte
}

func (c *stubVPNClient) GetDataKey(context.Context, *vpnproto.GetDataKeyRequest, *grpc.ClientConn) (*vpnproto.GetDataKeyResponse, error) {
	return &vpnproto.GetDataKeyResponse{DataKey: c.dataKey}, c.getDataKeyErr
}

func TestConstellationKMS(t *testing.T) {
	testCases := map[string]struct {
		vpn     *stubVPNClient
		wantErr bool
	}{
		"GetDataKey success": {
			vpn:     &stubVPNClient{dataKey: []byte{0x1, 0x2, 0x3}},
			wantErr: false,
		},
		"GetDataKey error": {
			vpn:     &stubVPNClient{getDataKeyErr: errors.New("error")},
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
				vpn:      tc.vpn,
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
