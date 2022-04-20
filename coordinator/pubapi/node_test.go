package pubapi

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestActivateAsNode(t *testing.T) {
	someErr := errors.New("failed")
	peer1 := peer.Peer{PublicIP: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}}
	peer2 := peer.Peer{PublicIP: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}}

	testCases := map[string]struct {
		initialPeers            []peer.Peer
		updatedPeers            []peer.Peer
		state                   state.State
		getUpdateErr            error
		setVPNIPErr             error
		messageSequenceOverride []string
		expectErr               bool
		expectedState           state.State
	}{
		"basic": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.AcceptingInit,
			expectedState: state.NodeWaitingForClusterJoin,
		},
		"already activated": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.IsNode,
			expectErr:     true,
			expectedState: state.IsNode,
		},
		"wrong peer kind": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.ActivatingNodes,
			expectErr:     true,
			expectedState: state.ActivatingNodes,
		},
		"GetUpdate error": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.AcceptingInit,
			getUpdateErr:  someErr,
			expectedState: state.NodeWaitingForClusterJoin,
		},
		"SetVPNIP error": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.AcceptingInit,
			setVPNIPErr:   someErr,
			expectErr:     true,
			expectedState: state.Failed,
		},
		"no messages sent to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{},
			expectErr:               true,
			expectedState:           state.AcceptingInit,
		},
		"only initial message sent to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{"initialRequest"},
			expectErr:               true,
			expectedState:           state.Failed,
		},
		"wrong initial message sent to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{"stateDiskKey"},
			expectErr:               true,
			expectedState:           state.AcceptingInit,
		},
		"initial message sent twice to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{"initialRequest", "initialRequest"},
			expectErr:               true,
			expectedState:           state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			const (
				nodeIP    = "192.0.2.2"
				nodeVPNIP = "10.118.0.2"
			)
			vpnPubKey := []byte{7, 8, 9}
			ownerID := []byte("ownerID")
			clusterID := []byte("clusterID")
			stateDiskKey := []byte("stateDiskKey")
			messageSequence := []string{"initialRequest", "stateDiskKey"}
			if tc.messageSequenceOverride != nil {
				messageSequence = tc.messageSequenceOverride
			}

			logger := zaptest.NewLogger(t)
			cor := &fakeCore{state: tc.state, vpnPubKey: vpnPubKey, setVPNIPErr: tc.setVPNIPErr}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, cor, dialer, nil, nil, nil)
			defer api.Close()

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{peers: tc.updatedPeers, getUpdateErr: tc.getUpdateErr}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(dialer.GetListener(net.JoinHostPort("10.118.0.1", vpnAPIPort)))
			defer vserver.GracefulStop()

			tlsConfig, err := atls.CreateAttestationServerTLSConfig(&core.MockIssuer{})
			require.NoError(err)
			pubserver := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
			pubproto.RegisterAPIServer(pubserver, api)
			go pubserver.Serve(dialer.GetListener(net.JoinHostPort(nodeIP, endpointAVPNPort)))
			defer pubserver.GracefulStop()

			_, nodeVPNPubKey, err := activateNode(require, dialer, messageSequence, nodeIP, "9000", nodeVPNIP, peer.ToPubProto(tc.initialPeers), ownerID, clusterID, stateDiskKey)
			assert.Equal(tc.expectedState, cor.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(vpnPubKey, nodeVPNPubKey)
			assert.Equal(nodeVPNIP, cor.vpnIP)
			assert.Equal(ownerID, cor.ownerID)
			assert.Equal(clusterID, cor.clusterID)

			api.Close() // blocks until update loop finished

			if tc.getUpdateErr == nil {
				require.Len(cor.updatedPeers, 2)
				assert.Equal(tc.updatedPeers, cor.updatedPeers[1])
			} else {
				require.Len(cor.updatedPeers, 1)
			}
			assert.Equal(tc.initialPeers, cor.updatedPeers[0])
			assert.Equal([]role.Role{role.Node}, cor.persistNodeStateRoles)
		})
	}
}

func TestTriggerNodeUpdate(t *testing.T) {
	someErr := errors.New("failed")
	peers := []peer.Peer{
		{PublicIP: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}},
		{PublicIP: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}},
	}

	testCases := map[string]struct {
		peers        []peer.Peer
		state        state.State
		getUpdateErr error
		expectErr    bool
	}{
		"basic": {
			peers: peers,
			state: state.IsNode,
		},
		"not activated": {
			peers:     peers,
			state:     state.AcceptingInit,
			expectErr: true,
		},
		"wrong peer kind": {
			peers:     peers,
			state:     state.ActivatingNodes,
			expectErr: true,
		},
		"GetUpdate error": {
			peers:        peers,
			state:        state.IsNode,
			getUpdateErr: someErr,
			expectErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, core, dialer, nil, nil, nil)

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{
				peers:        tc.peers,
				getUpdateErr: tc.getUpdateErr,
			}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(dialer.GetListener(net.JoinHostPort("10.118.0.1", vpnAPIPort)))
			defer vserver.GracefulStop()

			_, err := api.TriggerNodeUpdate(context.Background(), &pubproto.TriggerNodeUpdateRequest{})
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// second update should be a noop
			_, err = api.TriggerNodeUpdate(context.Background(), &pubproto.TriggerNodeUpdateRequest{})
			require.NoError(err)

			require.Len(core.updatedPeers, 1)
			assert.Equal(tc.peers, core.updatedPeers[0])
		})
	}
}

func TestJoinCluster(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		state          state.State
		getJoinArgsErr error
		joinClusterErr error
		expectErr      bool
		expectedState  state.State
	}{
		"basic": {
			state:         state.NodeWaitingForClusterJoin,
			expectedState: state.IsNode,
		},
		"not activated": {
			state:         state.AcceptingInit,
			expectErr:     true,
			expectedState: state.AcceptingInit,
		},
		"wrong peer kind": {
			state:         state.ActivatingNodes,
			expectErr:     true,
			expectedState: state.ActivatingNodes,
		},
		"GetK8sJoinArgs error": {
			state:          state.NodeWaitingForClusterJoin,
			getJoinArgsErr: someErr,
			expectErr:      true,
			expectedState:  state.NodeWaitingForClusterJoin,
		},
		"JoinCluster error": {
			state:          state.NodeWaitingForClusterJoin,
			joinClusterErr: someErr,
			expectErr:      true,
			expectedState:  state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state, joinClusterErr: tc.joinClusterErr}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, core, dialer, nil, nil, nil)

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{
				joinArgs: kubeadm.BootstrapTokenDiscovery{
					APIServerEndpoint: "endp",
					Token:             "token",
					CACertHashes:      []string{"dis"},
				},
				getJoinArgsErr: tc.getJoinArgsErr,
			}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(dialer.GetListener(net.JoinHostPort("192.0.2.1", vpnAPIPort)))
			defer vserver.GracefulStop()

			_, err := api.JoinCluster(context.Background(), &pubproto.JoinClusterRequest{CoordinatorVpnIp: "192.0.2.1"})

			assert.Equal(tc.expectedState, core.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal([]kubeadm.BootstrapTokenDiscovery{vapi.joinArgs}, core.joinArgs)
		})
	}
}

func activateNode(require *require.Assertions, dialer Dialer, messageSequence []string, nodeIP, bindPort, nodeVPNIP string, peers []*pubproto.Peer, ownerID, clusterID, stateDiskKey []byte) (string, []byte, error) {
	ctx := context.Background()
	conn, err := dialGRPC(ctx, dialer, net.JoinHostPort(nodeIP, bindPort))
	require.NoError(err)
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	stream, err := client.ActivateAsNode(ctx)
	if err != nil {
		return "", nil, err
	}

	for _, message := range messageSequence {
		switch message {
		case "initialRequest":
			err = stream.Send(&pubproto.ActivateAsNodeRequest{
				Request: &pubproto.ActivateAsNodeRequest_InitialRequest{
					InitialRequest: &pubproto.ActivateAsNodeInitialRequest{
						NodeVpnIp: nodeVPNIP,
						Peers:     peers,
						OwnerId:   ownerID,
						ClusterId: clusterID,
					},
				},
			})
			if err != nil {
				return "", nil, err
			}
		case "stateDiskKey":
			err = stream.Send(&pubproto.ActivateAsNodeRequest{
				Request: &pubproto.ActivateAsNodeRequest_StateDiskKey{
					StateDiskKey: stateDiskKey,
				},
			})
			if err != nil {
				return "", nil, err
			}
		default:
			panic("unknown message in activation")
		}
	}
	require.NoError(stream.CloseSend())

	diskUUIDReq, err := stream.Recv()
	if err != nil {
		return "", nil, err
	}
	diskUUID := diskUUIDReq.GetStateDiskUuid()

	vpnPubKeyReq, err := stream.Recv()
	if err != nil {
		return "", nil, err
	}
	nodeVPNPubKey := vpnPubKeyReq.GetNodeVpnPubKey()

	_, err = stream.Recv()
	if err != io.EOF {
		return "", nil, err
	}

	return diskUUID, nodeVPNPubKey, nil
}

func dialGRPC(ctx context.Context, dialer Dialer, target string) (*grpc.ClientConn, error) {
	tlsConfig, err := atls.CreateAttestationClientTLSConfig([]atls.Validator{&core.MockValidator{}})
	if err != nil {
		return nil, err
	}

	return grpc.DialContext(ctx, target,
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
}

type stubVPNAPI struct {
	peers          []peer.Peer
	getUpdateErr   error
	joinArgs       kubeadm.BootstrapTokenDiscovery
	getJoinArgsErr error
	vpnproto.UnimplementedAPIServer
}

func (a *stubVPNAPI) GetUpdate(ctx context.Context, in *vpnproto.GetUpdateRequest) (*vpnproto.GetUpdateResponse, error) {
	return &vpnproto.GetUpdateResponse{ResourceVersion: 1, Peers: peer.ToVPNProto(a.peers)}, a.getUpdateErr
}

func (a *stubVPNAPI) GetK8SJoinArgs(ctx context.Context, in *vpnproto.GetK8SJoinArgsRequest) (*vpnproto.GetK8SJoinArgsResponse, error) {
	return &vpnproto.GetK8SJoinArgsResponse{
		ApiServerEndpoint:        a.joinArgs.APIServerEndpoint,
		Token:                    a.joinArgs.Token,
		DiscoveryTokenCaCertHash: a.joinArgs.CACertHashes[0],
	}, a.getJoinArgsErr
}
