package pubapi

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetVPNPeers(t *testing.T) {
	wantedPeers := []peer.Peer{
		{
			PublicIP:  "192.0.2.1",
			VPNIP:     "10.118.0.1",
			VPNPubKey: []byte{0x1, 0x2, 0x3},
			Role:      role.Coordinator,
		},
	}

	testCases := map[string]struct {
		coreGetPeersErr error
		wantErr         bool
	}{
		"GetVPNPeers works": {},
		"GetVPNPeers fails if core cannot retrieve VPN peers": {
			coreGetPeersErr: errors.New("failed to get peers"),
			wantErr:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			cor := &fakeCore{peers: wantedPeers, GetPeersErr: tc.coreGetPeersErr}
			api := New(logger, cor, nil, nil, nil, nil)
			defer api.Close()
			resp, err := api.GetVPNPeers(context.Background(), &pubproto.GetVPNPeersRequest{})
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			peers := peer.FromPubProto(resp.Peers)
			assert.Equal(wantedPeers, peers)
		})
	}
}
