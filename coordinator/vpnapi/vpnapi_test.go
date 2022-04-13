package vpnapi

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap/zaptest"
	gpeer "google.golang.org/grpc/peer"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/kubernetes/klog/issues/282, https://github.com/kubernetes/klog/issues/188
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

func TestGetUpdate(t *testing.T) {
	someErr := errors.New("failed")
	clientIP := &net.IPAddr{IP: net.ParseIP("192.0.2.1")}
	peer1 := peer.Peer{PublicIP: "192.0.2.11", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}}
	peer2 := peer.Peer{PublicIP: "192.0.2.12", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}}
	peer3 := peer.Peer{PublicIP: "192.0.2.13", VPNIP: "192.0.2.23", VPNPubKey: []byte{3, 4, 5}}

	testCases := map[string]struct {
		clientAddr  net.Addr
		peers       []peer.Peer
		getPeersErr error
		expectErr   bool
	}{
		"0 peers": {
			clientAddr: clientIP,
			peers:      []peer.Peer{},
		},
		"1 peer": {
			clientAddr: clientIP,
			peers:      []peer.Peer{peer1},
		},
		"2 peers": {
			clientAddr: clientIP,
			peers:      []peer.Peer{peer1, peer2},
		},
		"3 peers": {
			clientAddr: clientIP,
			peers:      []peer.Peer{peer1, peer2, peer3},
		},
		"nil peers": {
			clientAddr: clientIP,
			peers:      nil,
		},
		"getPeers error": {
			clientAddr:  clientIP,
			getPeersErr: someErr,
			expectErr:   true,
		},
		"missing client addr": {
			peers: []peer.Peer{peer1},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			const serverResourceVersion = 2
			const clientResourceVersion = 3

			core := &stubCore{peers: tc.peers, serverResourceVersion: serverResourceVersion, getPeersErr: tc.getPeersErr}
			api := New(zaptest.NewLogger(t), core)

			ctx := context.Background()
			if tc.clientAddr != nil {
				ctx = gpeer.NewContext(ctx, &gpeer.Peer{Addr: tc.clientAddr})
			}

			resp, err := api.GetUpdate(ctx, &vpnproto.GetUpdateRequest{ResourceVersion: clientResourceVersion})
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.EqualValues(serverResourceVersion, resp.ResourceVersion)
			assert.Equal([]int{clientResourceVersion}, core.clientResourceVersions)

			require.Len(resp.Peers, len(tc.peers))
			for i, actual := range resp.Peers {
				expected := tc.peers[i]
				assert.EqualValues(expected.PublicIP, actual.PublicIp)
				assert.EqualValues(expected.VPNIP, actual.VpnIp)
				assert.Equal(expected.VPNPubKey, actual.VpnPubKey)
			}

			if tc.clientAddr == nil {
				assert.Empty(core.heartbeats)
			} else {
				assert.Equal([]net.Addr{tc.clientAddr}, core.heartbeats)
			}
		})
	}
}

func TestGetK8SJoinArgs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	joinArgs := kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "endp",
		Token:             "token",
		CACertHashes:      []string{"dis"},
	}
	api := New(zaptest.NewLogger(t), &stubCore{joinArgs: joinArgs})

	resp, err := api.GetK8SJoinArgs(context.Background(), &vpnproto.GetK8SJoinArgsRequest{})
	require.NoError(err)
	assert.Equal(joinArgs.APIServerEndpoint, resp.ApiServerEndpoint)
	assert.Equal(joinArgs.Token, resp.Token)
	assert.Equal(joinArgs.CACertHashes[0], resp.DiscoveryTokenCaCertHash)
}

func TestGetDataKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	core := &stubCore{derivedKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5}}
	api := New(zaptest.NewLogger(t), core)
	res, err := api.GetDataKey(context.Background(), &vpnproto.GetDataKeyRequest{DataKeyId: "key-1", Length: 32})
	require.NoError(err)
	assert.Equal(core.derivedKey, res.DataKey)

	api = New(zaptest.NewLogger(t), &stubCore{deriveKeyErr: errors.New("error")})
	res, err = api.GetDataKey(context.Background(), &vpnproto.GetDataKeyRequest{DataKeyId: "key-1", Length: 32})
	assert.Error(err)
	assert.Nil(res)
}

type stubCore struct {
	peers                  []peer.Peer
	serverResourceVersion  int
	getPeersErr            error
	clientResourceVersions []int
	heartbeats             []net.Addr
	joinArgs               kubeadm.BootstrapTokenDiscovery
	derivedKey             []byte
	deriveKeyErr           error
}

func (c *stubCore) GetPeers(resourceVersion int) (int, []peer.Peer, error) {
	c.clientResourceVersions = append(c.clientResourceVersions, resourceVersion)
	return c.serverResourceVersion, c.peers, c.getPeersErr
}

func (c *stubCore) NotifyNodeHeartbeat(addr net.Addr) {
	c.heartbeats = append(c.heartbeats, addr)
}

func (c *stubCore) GetK8sJoinArgs() (*kubeadm.BootstrapTokenDiscovery, error) {
	return &c.joinArgs, nil
}

func (c *stubCore) GetDataKey(ctx context.Context, dataKeyID string, length int) ([]byte, error) {
	if c.deriveKeyErr != nil {
		return nil, c.deriveKeyErr
	}
	return c.derivedKey, nil
}
