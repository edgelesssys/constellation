package pubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/logging"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/oid"
	kms "github.com/edgelesssys/constellation/kms/server/setup"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	grpcpeer "google.golang.org/grpc/peer"
)

func TestActivateAsCoordinator(t *testing.T) {
	someErr := errors.New("failed")
	coordinatorPubKey := []byte{6, 7, 8}
	testNode1 := newStubPeer("192.0.2.11", []byte{1, 2, 3})
	testNode2 := newStubPeer("192.0.2.12", []byte{2, 3, 4})
	testNode3 := newStubPeer("192.0.2.13", []byte{3, 4, 5})
	wantNode1 := peer.Peer{PublicIP: "192.0.2.11", VPNIP: "10.118.0.11", VPNPubKey: []byte{1, 2, 3}, Role: role.Node}
	wantNode2 := peer.Peer{PublicIP: "192.0.2.12", VPNIP: "10.118.0.12", VPNPubKey: []byte{2, 3, 4}, Role: role.Node}
	wantNode3 := peer.Peer{PublicIP: "192.0.2.13", VPNIP: "10.118.0.13", VPNPubKey: []byte{3, 4, 5}, Role: role.Node}
	wantCoord := peer.Peer{PublicIP: "192.0.2.1", VPNIP: "10.118.0.1", VPNPubKey: coordinatorPubKey, Role: role.Coordinator}
	adminPeer := peer.Peer{VPNPubKey: []byte{7, 8, 9}, Role: role.Admin}
	sshUser1 := &ssh.UserKey{
		Username:  "test-user-1",
		PublicKey: "ssh-rsa abcdefg",
	}
	sshUser2 := &ssh.UserKey{
		Username:  "test-user-2",
		PublicKey: "ssh-ed25519 hijklmn",
	}

	testCases := map[string]struct {
		nodes                      []*stubPeer
		state                      state.State
		switchToPersistentStoreErr error
		wantErr                    bool
		wantPeers                  []peer.Peer
		wantState                  state.State
		adminVPNIP                 string
		sshKeys                    []*ssh.UserKey
	}{
		"0 nodes": {
			state:      state.AcceptingInit,
			wantPeers:  []peer.Peer{wantCoord},
			wantState:  state.ActivatingNodes,
			adminVPNIP: "10.118.0.11",
		},
		"1 node": {
			nodes:      []*stubPeer{testNode1},
			state:      state.AcceptingInit,
			wantPeers:  []peer.Peer{wantCoord, wantNode1},
			wantState:  state.ActivatingNodes,
			adminVPNIP: "10.118.0.12",
		},
		"2 nodes": {
			nodes:      []*stubPeer{testNode1, testNode2},
			state:      state.AcceptingInit,
			wantPeers:  []peer.Peer{wantCoord, wantNode1, wantNode2},
			wantState:  state.ActivatingNodes,
			adminVPNIP: "10.118.0.13",
		},
		"3 nodes": {
			nodes:      []*stubPeer{testNode1, testNode2, testNode3},
			state:      state.AcceptingInit,
			wantPeers:  []peer.Peer{wantCoord, wantNode1, wantNode2, wantNode3},
			wantState:  state.ActivatingNodes,
			adminVPNIP: "10.118.0.14",
		},
		"coordinator with SSH users": {
			state:      state.AcceptingInit,
			wantPeers:  []peer.Peer{wantCoord},
			wantState:  state.ActivatingNodes,
			adminVPNIP: "10.118.0.11",
			sshKeys:    []*ssh.UserKey{sshUser1, sshUser2},
		},
		"already activated": {
			nodes:     []*stubPeer{testNode1},
			state:     state.ActivatingNodes,
			wantErr:   true,
			wantState: state.ActivatingNodes,
		},
		"wrong peer kind": {
			nodes:     []*stubPeer{testNode1},
			state:     state.IsNode,
			wantErr:   true,
			wantState: state.IsNode,
		},
		"node activation error": {
			nodes:     []*stubPeer{testNode1, {activateErr: someErr}, testNode3},
			state:     state.AcceptingInit,
			wantErr:   true,
			wantState: state.Failed,
		},
		"node join error": {
			nodes:     []*stubPeer{testNode1, {joinErr: someErr}, testNode3},
			state:     state.AcceptingInit,
			wantErr:   true,
			wantState: state.Failed,
		},
		"SwitchToPersistentStore error": {
			nodes:                      []*stubPeer{testNode1},
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			wantErr:                    true,
			wantState:                  state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			autoscalingNodeGroups := []string{"ang1", "ang2"}
			keyEncryptionKeyID := "constellation"
			fs := afero.NewMemMapFs()
			core := &fakeCore{
				state:                      tc.state,
				vpnPubKey:                  coordinatorPubKey,
				switchToPersistentStoreErr: tc.switchToPersistentStoreErr,
				kubeconfig:                 []byte("kubeconfig"),
				ownerID:                    []byte("ownerID"),
				clusterID:                  []byte("clusterID"),
				linuxUserManager:           user.NewLinuxUserManagerFake(fs),
			}

			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, fakeValidator{}, netDialer)

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, dialer, stubVPNAPIServer{}, getPublicIPAddr, nil)
			defer api.Close()

			// spawn nodes
			var nodePublicIPs []string
			var wg sync.WaitGroup
			for _, n := range tc.nodes {
				nodePublicIPs = append(nodePublicIPs, n.peer.PublicIP)
				server := n.newServer()
				wg.Add(1)
				go func(endpoint string) {
					listener := netDialer.GetListener(endpoint)
					wg.Done()
					_ = server.Serve(listener)
				}(net.JoinHostPort(n.peer.PublicIP, endpointAVPNPort))
				defer server.GracefulStop()
			}
			wg.Wait()

			stream := &stubActivateAsCoordinatorServer{}
			err := api.ActivateAsCoordinator(&pubproto.ActivateAsCoordinatorRequest{
				AdminVpnPubKey:        adminPeer.VPNPubKey,
				NodePublicIps:         nodePublicIPs,
				AutoscalingNodeGroups: autoscalingNodeGroups,
				MasterSecret:          []byte("Constellation"),
				KeyEncryptionKeyId:    keyEncryptionKeyID,
				UseExistingKek:        false,
				KmsUri:                kms.ClusterKMSURI,
				StorageUri:            kms.NoStoreURI,
				SshUserKeys:           ssh.ToProtoSlice(tc.sshKeys),
			}, stream)

			assert.Equal(tc.wantState, core.state)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// Coordinator streams logs and admin conf
			require.Greater(len(stream.sent), len(tc.nodes))
			for i := 0; i < len(stream.sent)-1; i++ {
				assert.NotEmpty(stream.sent[i].GetLog().Message)
			}
			adminConfig := stream.sent[len(stream.sent)-1].GetAdminConfig()
			assert.Equal(tc.adminVPNIP, adminConfig.AdminVpnIp)
			assert.Equal(coordinatorPubKey, adminConfig.CoordinatorVpnPubKey)
			assert.Equal(core.kubeconfig, adminConfig.Kubeconfig)
			assert.Equal(core.ownerID, adminConfig.OwnerId)
			assert.Equal(core.clusterID, adminConfig.ClusterId)

			// Core is updated
			vpnIP, err := core.GetVPNIP()
			require.NoError(err)
			assert.Equal(vpnIP, core.vpnIP)
			// construct full list of expected peers
			adminPeer.VPNIP = tc.adminVPNIP
			assert.Equal(append(tc.wantPeers, adminPeer), core.peers)
			assert.Equal(autoscalingNodeGroups, core.autoscalingNodeGroups)
			assert.Equal(keyEncryptionKeyID, core.kekID)
			assert.Equal([]role.Role{role.Coordinator}, core.persistNodeStateRoles)

			// Test SSH user & key creation. Both cases: "supposed to add" and "not supposed to add"
			// This slightly differs from a real environment (e.g. missing /var/home) but should be fine in the stub context with a virtual file system
			if tc.sshKeys != nil {
				passwd := user.Passwd{}
				entries, err := passwd.Parse(fs)
				require.NoError(err)
				for _, singleEntry := range entries {
					username := singleEntry.Gecos
					_, err := fs.Stat(fmt.Sprintf("/var/home/%s/.ssh/authorized_keys.d/constellation-ssh-keys", username))
					assert.NoError(err)
				}
			} else {
				passwd := user.Passwd{}
				_, err := passwd.Parse(fs)
				assert.EqualError(err, "open /etc/passwd: file does not exist")
				_, err = fs.Stat("/var/home")
				assert.EqualError(err, "open /var/home: file does not exist")
			}
		})
	}
}

func TestActivateAdditionalNodes(t *testing.T) {
	someErr := errors.New("failed")
	testNode1 := newStubPeer("192.0.2.11", []byte{1, 2, 3})
	testNode2 := newStubPeer("192.0.2.12", []byte{2, 3, 4})
	testNode3 := newStubPeer("192.0.2.13", []byte{3, 4, 5})
	wantNode1 := peer.Peer{PublicIP: "192.0.2.11", VPNIP: "10.118.0.11", VPNPubKey: []byte{1, 2, 3}, Role: role.Node}
	wantNode2 := peer.Peer{PublicIP: "192.0.2.12", VPNIP: "10.118.0.12", VPNPubKey: []byte{2, 3, 4}, Role: role.Node}
	wantNode3 := peer.Peer{PublicIP: "192.0.2.13", VPNIP: "10.118.0.13", VPNPubKey: []byte{3, 4, 5}, Role: role.Node}

	testCases := map[string]struct {
		nodes     []*stubPeer
		state     state.State
		wantErr   bool
		wantPeers []peer.Peer
	}{
		"0 nodes": {
			state: state.ActivatingNodes,
		},
		"1 node": {
			nodes:     []*stubPeer{testNode1},
			state:     state.ActivatingNodes,
			wantPeers: []peer.Peer{wantNode1},
		},
		"2 nodes": {
			nodes:     []*stubPeer{testNode1, testNode2},
			state:     state.ActivatingNodes,
			wantPeers: []peer.Peer{wantNode1, wantNode2},
		},
		"3 nodes": {
			nodes:     []*stubPeer{testNode1, testNode2, testNode3},
			state:     state.ActivatingNodes,
			wantPeers: []peer.Peer{wantNode1, wantNode2, wantNode3},
		},
		"uninitialized": {
			nodes:   []*stubPeer{testNode1},
			wantErr: true,
		},
		"wrong peer kind": {
			nodes:   []*stubPeer{testNode1},
			state:   state.IsNode,
			wantErr: true,
		},
		"node activation error": {
			nodes:   []*stubPeer{testNode1, {activateErr: someErr}, testNode3},
			state:   state.ActivatingNodes,
			wantErr: true,
		},
		"node join error": {
			nodes:   []*stubPeer{testNode1, {joinErr: someErr}, testNode3},
			state:   state.ActivatingNodes,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core := &fakeCore{state: tc.state}
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, fakeValidator{}, netDialer)

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, dialer, nil, getPublicIPAddr, nil)
			defer api.Close()
			// spawn nodes
			var nodePublicIPs []string
			var wg sync.WaitGroup
			for _, n := range tc.nodes {
				nodePublicIPs = append(nodePublicIPs, n.peer.PublicIP)
				server := n.newServer()
				wg.Add(1)
				go func(endpoint string) {
					listener := netDialer.GetListener(endpoint)
					wg.Done()
					_ = server.Serve(listener)
				}(net.JoinHostPort(n.peer.PublicIP, endpointAVPNPort))
				defer server.GracefulStop()
			}
			wg.Wait()
			// since we are not activating the coordinator, initialize the store with IP's
			require.NoError(core.InitializeStoreIPs())
			stream := &stubActivateAdditionalNodesServer{}
			err := api.ActivateAdditionalNodes(&pubproto.ActivateAdditionalNodesRequest{NodePublicIps: nodePublicIPs}, stream)
			if tc.wantErr {
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
			assert.Equal(tc.wantPeers, core.peers)
		})
	}
}

func TestAssemblePeerStruct(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	getPublicIPAddr := func() (string, error) {
		return "192.0.2.1", nil
	}

	vpnPubKey := []byte{2, 3, 4}
	core := &fakeCore{vpnPubKey: vpnPubKey}
	api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, nil, nil, getPublicIPAddr, nil)
	defer api.Close()

	vpnIP, err := core.GetVPNIP()
	require.NoError(err)
	want := peer.Peer{
		PublicIP:  "192.0.2.1",
		VPNIP:     vpnIP,
		VPNPubKey: vpnPubKey,
		Role:      role.Coordinator,
	}

	actual, err := api.assemblePeerStruct(vpnIP, role.Coordinator)
	require.NoError(err)
	assert.Equal(want, actual)
}

type stubPeer struct {
	peer                   peer.Peer
	activateAsNodeMessages []*pubproto.ActivateAsNodeResponse
	activateAsNodeReceive  int
	activateErr            error
	joinErr                error
	getPubKeyErr           error
	pubproto.UnimplementedAPIServer
}

func newStubPeer(publicIP string, vpnPubKey []byte) *stubPeer {
	return &stubPeer{
		peer: peer.Peer{PublicIP: publicIP, VPNPubKey: vpnPubKey},
		activateAsNodeMessages: []*pubproto.ActivateAsNodeResponse{
			{Response: &pubproto.ActivateAsNodeResponse_StateDiskUuid{StateDiskUuid: "state-disk-uuid"}},
			{Response: &pubproto.ActivateAsNodeResponse_NodeVpnPubKey{NodeVpnPubKey: vpnPubKey}},
		},
		activateAsNodeReceive: 2,
	}
}

func (n *stubPeer) ActivateAsNode(stream pubproto.API_ActivateAsNodeServer) error {
	for _, message := range n.activateAsNodeMessages {
		err := stream.Send(message)
		if err != nil {
			return err
		}
	}
	for i := 0; i < n.activateAsNodeReceive; i++ {
		_, err := stream.Recv()
		if err != nil {
			return err
		}
	}
	if _, err := stream.Recv(); err != io.EOF {
		return err
	}

	return n.activateErr
}

func (n *stubPeer) ActivateAsAdditionalCoordinator(ctx context.Context, in *pubproto.ActivateAsAdditionalCoordinatorRequest) (*pubproto.ActivateAsAdditionalCoordinatorResponse, error) {
	return &pubproto.ActivateAsAdditionalCoordinatorResponse{}, n.activateErr
}

func (*stubPeer) TriggerNodeUpdate(ctx context.Context, in *pubproto.TriggerNodeUpdateRequest) (*pubproto.TriggerNodeUpdateResponse, error) {
	return &pubproto.TriggerNodeUpdateResponse{}, nil
}

func (n *stubPeer) JoinCluster(ctx context.Context, in *pubproto.JoinClusterRequest) (*pubproto.JoinClusterResponse, error) {
	return &pubproto.JoinClusterResponse{}, n.joinErr
}

func (n *stubPeer) GetPeerVPNPublicKey(ctx context.Context, in *pubproto.GetPeerVPNPublicKeyRequest) (*pubproto.GetPeerVPNPublicKeyResponse, error) {
	return &pubproto.GetPeerVPNPublicKeyResponse{CoordinatorPubKey: n.peer.VPNPubKey}, n.getPubKeyErr
}

func (n *stubPeer) newServer() *grpc.Server {
	creds := atlscredentials.New(fakeIssuer{}, nil)
	server := grpc.NewServer(grpc.Creds(creds))
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

func TestRequestStateDiskKey(t *testing.T) {
	defaultKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	someErr := errors.New("error")
	testCases := map[string]struct {
		state         state.State
		dataKey       []byte
		getDataKeyErr error
		pushKeyErr    error
		wantErr       bool
	}{
		"success": {
			state:   state.ActivatingNodes,
			dataKey: defaultKey,
		},
		"Coordinator in wrong state": {
			state:   state.IsNode,
			dataKey: defaultKey,
			wantErr: true,
		},
		"GetDataKey fails": {
			state:         state.ActivatingNodes,
			dataKey:       defaultKey,
			getDataKeyErr: someErr,
			wantErr:       true,
		},
		"key pushing fails": {
			state:      state.ActivatingNodes,
			dataKey:    defaultKey,
			pushKeyErr: someErr,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			issuer := atls.NewFakeIssuer(oid.Dummy{})

			stateDiskServer := &stubStateDiskServer{pushKeyErr: tc.pushKeyErr}

			// we can not use a bufconn here, since we rely on grpcpeer.FromContext() to connect to the caller
			listener, err := net.Listen("tcp", ":")
			require.NoError(err)
			defer listener.Close()

			creds := atlscredentials.New(issuer, nil)
			s := grpc.NewServer(grpc.Creds(creds))
			keyproto.RegisterAPIServer(s, stateDiskServer)
			defer s.GracefulStop()
			go s.Serve(listener)

			ctx := grpcpeer.NewContext(context.Background(), &grpcpeer.Peer{Addr: listener.Addr()})
			getPeerFromContext := func(ctx context.Context) (string, error) {
				peer, ok := grpcpeer.FromContext(ctx)
				if !ok {
					return "", errors.New("unable to get peer from context")
				}
				return peer.Addr.String(), nil
			}

			core := &fakeCore{
				state:         tc.state,
				dataKey:       tc.dataKey,
				getDataKeyErr: tc.getDataKeyErr,
			}

			api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, dialer.New(nil, dummyValidator{}, &net.Dialer{}), nil, nil, getPeerFromContext)

			_, err = api.RequestStateDiskKey(ctx, &pubproto.RequestStateDiskKeyRequest{})
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.dataKey, stateDiskServer.receivedRequest.StateDiskKey)
			}
		})
	}
}

type dummyValidator struct {
	oid.Dummy
}

func (d dummyValidator) Validate(attdoc []byte, nonce []byte) ([]byte, error) {
	var attestation vtpm.AttestationDocument
	if err := json.Unmarshal(attdoc, &attestation); err != nil {
		return nil, err
	}
	return attestation.UserData, nil
}

type stubStateDiskServer struct {
	receivedRequest *keyproto.PushStateDiskKeyRequest
	pushKeyErr      error
	keyproto.UnimplementedAPIServer
}

func (s *stubStateDiskServer) PushStateDiskKey(ctx context.Context, in *keyproto.PushStateDiskKeyRequest) (*keyproto.PushStateDiskKeyResponse, error) {
	s.receivedRequest = in
	return &keyproto.PushStateDiskKeyResponse{}, s.pushKeyErr
}
