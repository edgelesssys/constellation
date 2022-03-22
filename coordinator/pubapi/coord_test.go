package pubapi

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/oid"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestActivateAsCoordinator(t *testing.T) {
	someErr := errors.New("failed")
	coordinatorPubKey := []byte{6, 7, 8}
	testNode1 := &stubNode{publicIP: "192.0.2.11", pubKey: []byte{1, 2, 3}}
	testNode2 := &stubNode{publicIP: "192.0.2.12", pubKey: []byte{2, 3, 4}}
	testNode3 := &stubNode{publicIP: "192.0.2.13", pubKey: []byte{3, 4, 5}}
	expectedNode1 := peer.Peer{PublicEndpoint: "192.0.2.11:9000", VPNIP: "192.0.2.101", VPNPubKey: []byte{1, 2, 3}, Role: role.Node}
	expectedNode2 := peer.Peer{PublicEndpoint: "192.0.2.12:9000", VPNIP: "192.0.2.102", VPNPubKey: []byte{2, 3, 4}, Role: role.Node}
	expectedNode3 := peer.Peer{PublicEndpoint: "192.0.2.13:9000", VPNIP: "192.0.2.103", VPNPubKey: []byte{3, 4, 5}, Role: role.Node}
	expectedCoord := peer.Peer{PublicEndpoint: "192.0.2.1:9000", VPNIP: "192.0.2.100", VPNPubKey: coordinatorPubKey, Role: role.Coordinator}

	testCases := map[string]struct {
		nodes                      []*stubNode
		state                      state.State
		switchToPersistentStoreErr error
		expectErr                  bool
		expectedPeers              []peer.Peer
		expectedState              state.State
	}{
		"0 nodes": {
			state:         state.AcceptingInit,
			expectedPeers: []peer.Peer{expectedCoord},
			expectedState: state.ActivatingNodes,
		},
		"1 node": {
			nodes:         []*stubNode{testNode1},
			state:         state.AcceptingInit,
			expectedPeers: []peer.Peer{expectedCoord, expectedNode1},
			expectedState: state.ActivatingNodes,
		},
		"2 nodes": {
			nodes:         []*stubNode{testNode1, testNode2},
			state:         state.AcceptingInit,
			expectedPeers: []peer.Peer{expectedCoord, expectedNode1, expectedNode2},
			expectedState: state.ActivatingNodes,
		},
		"3 nodes": {
			nodes:         []*stubNode{testNode1, testNode2, testNode3},
			state:         state.AcceptingInit,
			expectedPeers: []peer.Peer{expectedCoord, expectedNode1, expectedNode2, expectedNode3},
			expectedState: state.ActivatingNodes,
		},
		"already activated": {
			nodes:         []*stubNode{testNode1},
			state:         state.ActivatingNodes,
			expectErr:     true,
			expectedState: state.ActivatingNodes,
		},
		"wrong peer kind": {
			nodes:         []*stubNode{testNode1},
			state:         state.IsNode,
			expectErr:     true,
			expectedState: state.IsNode,
		},
		"node activation error": {
			nodes:         []*stubNode{testNode1, {activateErr: someErr}, testNode3},
			state:         state.AcceptingInit,
			expectErr:     true,
			expectedState: state.Failed,
		},
		"node join error": {
			nodes:         []*stubNode{testNode1, {joinErr: someErr}, testNode3},
			state:         state.AcceptingInit,
			expectErr:     true,
			expectedState: state.Failed,
		},
		"SwitchToPersistentStore error": {
			nodes:                      []*stubNode{testNode1},
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			expectErr:                  true,
			expectedState:              state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			adminPubKey := []byte{7, 8, 9}
			autoscalingNodeGroups := []string{"ang1", "ang2"}
			keyEncryptionKeyID := "constellation"

			core := &fakeCore{
				state:                      tc.state,
				vpnPubKey:                  coordinatorPubKey,
				switchToPersistentStoreErr: tc.switchToPersistentStoreErr,
				kubeconfig:                 []byte("kubeconfig"),
				ownerID:                    []byte("ownerID"),
				clusterID:                  []byte("clusterID"),
			}
			dialer := testdialer.NewBufconnDialer()

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), core, dialer, stubVPNAPIServer{}, fakeValidator{}, getPublicIPAddr)

			// spawn nodes
			var nodePublicEndpoints []string
			for _, n := range tc.nodes {
				publicEndpoint := net.JoinHostPort(n.publicIP, endpointAVPNPort)
				nodePublicEndpoints = append(nodePublicEndpoints, publicEndpoint)
				server := n.newServer()
				go server.Serve(dialer.GetListener(publicEndpoint))
				defer server.GracefulStop()
			}

			stream := &stubActivateAsCoordinatorServer{}
			err := api.ActivateAsCoordinator(&pubproto.ActivateAsCoordinatorRequest{
				AdminVpnPubKey:        adminPubKey,
				NodePublicEndpoints:   nodePublicEndpoints,
				AutoscalingNodeGroups: autoscalingNodeGroups,
				MasterSecret:          []byte("Constellation"),
				KeyEncryptionKeyId:    keyEncryptionKeyID,
				UseExistingKek:        false,
				KmsUri:                kms.ClusterKMSURI,
				StorageUri:            kms.NoStoreURI,
			}, stream)

			assert.Equal(tc.expectedState, core.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// Coordinator streams logs and admin conf
			require.Len(stream.sent, len(tc.nodes)+1)
			for i := 0; i < len(tc.nodes); i++ {
				assert.NotEmpty(stream.sent[i].GetLog().Message)
			}
			adminConfig := stream.sent[len(tc.nodes)].GetAdminConfig()
			assert.Equal("192.0.2.99", adminConfig.AdminVpnIp)
			assert.Equal(coordinatorPubKey, adminConfig.CoordinatorVpnPubKey)
			assert.Equal(core.kubeconfig, adminConfig.Kubeconfig)
			assert.Equal(core.ownerID, adminConfig.OwnerId)
			assert.Equal(core.clusterID, adminConfig.ClusterId)

			// Core is updated
			assert.Equal(adminPubKey, core.adminPubKey)
			assert.Equal(core.GetCoordinatorVPNIP(), core.vpnIP)
			assert.Equal(tc.expectedPeers, core.peers)
			assert.Equal(autoscalingNodeGroups, core.autoscalingNodeGroups)
			assert.Equal(keyEncryptionKeyID, core.kekID)
		})
	}
}

func TestActivateAdditionalNodes(t *testing.T) {
	someErr := errors.New("failed")
	testNode1 := &stubNode{publicIP: "192.0.2.11", pubKey: []byte{1, 2, 3}}
	testNode2 := &stubNode{publicIP: "192.0.2.12", pubKey: []byte{2, 3, 4}}
	testNode3 := &stubNode{publicIP: "192.0.2.13", pubKey: []byte{3, 4, 5}}
	expectedNode1 := peer.Peer{PublicEndpoint: "192.0.2.11:9000", VPNIP: "192.0.2.101", VPNPubKey: []byte{1, 2, 3}, Role: role.Node}
	expectedNode2 := peer.Peer{PublicEndpoint: "192.0.2.12:9000", VPNIP: "192.0.2.102", VPNPubKey: []byte{2, 3, 4}, Role: role.Node}
	expectedNode3 := peer.Peer{PublicEndpoint: "192.0.2.13:9000", VPNIP: "192.0.2.103", VPNPubKey: []byte{3, 4, 5}, Role: role.Node}

	testCases := map[string]struct {
		nodes         []*stubNode
		state         state.State
		expectErr     bool
		expectedPeers []peer.Peer
	}{
		"0 nodes": {
			state: state.ActivatingNodes,
		},
		"1 node": {
			nodes:         []*stubNode{testNode1},
			state:         state.ActivatingNodes,
			expectedPeers: []peer.Peer{expectedNode1},
		},
		"2 nodes": {
			nodes:         []*stubNode{testNode1, testNode2},
			state:         state.ActivatingNodes,
			expectedPeers: []peer.Peer{expectedNode1, expectedNode2},
		},
		"3 nodes": {
			nodes:         []*stubNode{testNode1, testNode2, testNode3},
			state:         state.ActivatingNodes,
			expectedPeers: []peer.Peer{expectedNode1, expectedNode2, expectedNode3},
		},
		"uninitialized": {
			nodes:     []*stubNode{testNode1},
			expectErr: true,
		},
		"wrong peer kind": {
			nodes:     []*stubNode{testNode1},
			state:     state.IsNode,
			expectErr: true,
		},
		"node activation error": {
			nodes:     []*stubNode{testNode1, {activateErr: someErr}, testNode3},
			state:     state.ActivatingNodes,
			expectErr: true,
		},
		"node join error": {
			nodes:     []*stubNode{testNode1, {joinErr: someErr}, testNode3},
			state:     state.ActivatingNodes,
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core := &fakeCore{state: tc.state}
			dialer := testdialer.NewBufconnDialer()

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), core, dialer, nil, fakeValidator{}, getPublicIPAddr)

			// spawn nodes
			var nodePublicEndpoints []string
			for _, n := range tc.nodes {
				publicEndpoint := net.JoinHostPort(n.publicIP, endpointAVPNPort)
				nodePublicEndpoints = append(nodePublicEndpoints, publicEndpoint)
				server := n.newServer()
				go server.Serve(dialer.GetListener(publicEndpoint))
				defer server.GracefulStop()
			}

			stream := &stubActivateAdditionalNodesServer{}
			err := api.ActivateAdditionalNodes(&pubproto.ActivateAdditionalNodesRequest{NodePublicEndpoints: nodePublicEndpoints}, stream)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// Coordinator streams logs
			require.Len(stream.sent, len(tc.nodes)+1)
			for _, s := range stream.sent {
				assert.NotEmpty(s.GetLog().Message)
			}

			// Core is updated
			assert.Equal(tc.expectedPeers, core.peers)
		})
	}
}

func TestMakeCoordinatorPeer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	getPublicIPAddr := func() (string, error) {
		return "192.0.2.1", nil
	}

	vpnPubKey := []byte{2, 3, 4}
	core := &fakeCore{vpnPubKey: vpnPubKey}
	api := New(zaptest.NewLogger(t), core, nil, nil, nil, getPublicIPAddr)

	expected := peer.Peer{
		PublicEndpoint: "192.0.2.1:9000",
		VPNIP:          core.GetCoordinatorVPNIP(),
		VPNPubKey:      vpnPubKey,
		Role:           role.Coordinator,
	}

	actual, err := api.makeCoordinatorPeer()
	require.NoError(err)
	assert.Equal(expected, actual)
}

type stubNode struct {
	publicIP    string
	pubKey      []byte
	activateErr error
	joinErr     error
	pubproto.UnimplementedAPIServer
}

func (n *stubNode) ActivateAsNode(ctx context.Context, in *pubproto.ActivateAsNodeRequest) (*pubproto.ActivateAsNodeResponse, error) {
	return &pubproto.ActivateAsNodeResponse{NodeVpnPubKey: n.pubKey}, n.activateErr
}

func (*stubNode) TriggerNodeUpdate(ctx context.Context, in *pubproto.TriggerNodeUpdateRequest) (*pubproto.TriggerNodeUpdateResponse, error) {
	return &pubproto.TriggerNodeUpdateResponse{}, nil
}

func (n *stubNode) JoinCluster(ctx context.Context, in *pubproto.JoinClusterRequest) (*pubproto.JoinClusterResponse, error) {
	return &pubproto.JoinClusterResponse{}, n.joinErr
}

func (n *stubNode) newServer() *grpc.Server {
	tlsConfig, err := atls.CreateAttestationServerTLSConfig(fakeIssuer{})
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pubproto.RegisterAPIServer(server, n)
	return server
}

type stubVPNAPIServer struct{}

func (stubVPNAPIServer) Listen(endpoint string) error {
	return nil
}

func (stubVPNAPIServer) Serve() error {
	return nil
}

func (stubVPNAPIServer) Close() {
}

type fakeIssuer struct {
	oid.Dummy
}

func (fakeIssuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	return userData, nil
}

type fakeValidator struct {
	oid.Dummy
}

func (fakeValidator) Validate(attdoc []byte, nonce []byte) ([]byte, error) {
	return attdoc, nil
}

type stubActivateAsCoordinatorServer struct {
	grpc.ServerStream
	sent []*pubproto.ActivateAsCoordinatorResponse
}

func (s *stubActivateAsCoordinatorServer) Send(req *pubproto.ActivateAsCoordinatorResponse) error {
	s.sent = append(s.sent, req)
	return nil
}

type stubActivateAdditionalNodesServer struct {
	grpc.ServerStream
	sent []*pubproto.ActivateAdditionalNodesResponse
}

func (s *stubActivateAdditionalNodesServer) Send(req *pubproto.ActivateAdditionalNodesResponse) error {
	s.sent = append(s.sent, req)
	return nil
}
