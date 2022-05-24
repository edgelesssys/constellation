package pubapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/util/grpcutil"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/spf13/afero"
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
	sshUser1 := &ssh.UserKey{
		Username:  "test-user-1",
		PublicKey: "ssh-rsa abcdefg",
	}
	sshUser2 := &ssh.UserKey{
		Username:  "test-user-2",
		PublicKey: "ssh-ed25519 hijklmn",
	}

	testCases := map[string]struct {
		initialPeers            []peer.Peer
		updatedPeers            []peer.Peer
		state                   state.State
		getUpdateErr            error
		setVPNIPErr             error
		messageSequenceOverride []string
		wantErr                 bool
		wantState               state.State
		sshKeys                 []*ssh.UserKey
	}{
		"basic": {
			initialPeers: []peer.Peer{peer1},
			updatedPeers: []peer.Peer{peer2},
			state:        state.AcceptingInit,
			wantState:    state.NodeWaitingForClusterJoin,
		},
		"basic with SSH users": {
			initialPeers: []peer.Peer{peer1},
			updatedPeers: []peer.Peer{peer2},
			state:        state.AcceptingInit,
			wantState:    state.NodeWaitingForClusterJoin,
			sshKeys:      []*ssh.UserKey{sshUser1, sshUser2},
		},
		"already activated": {
			initialPeers: []peer.Peer{peer1},
			updatedPeers: []peer.Peer{peer2},
			state:        state.IsNode,
			wantErr:      true,
			wantState:    state.IsNode,
		},
		"wrong peer kind": {
			initialPeers: []peer.Peer{peer1},
			updatedPeers: []peer.Peer{peer2},
			state:        state.ActivatingNodes,
			wantErr:      true,
			wantState:    state.ActivatingNodes,
		},
		"GetUpdate error": {
			initialPeers: []peer.Peer{peer1},
			updatedPeers: []peer.Peer{peer2},
			state:        state.AcceptingInit,
			getUpdateErr: someErr,
			wantState:    state.NodeWaitingForClusterJoin,
		},
		"SetVPNIP error": {
			initialPeers: []peer.Peer{peer1},
			updatedPeers: []peer.Peer{peer2},
			state:        state.AcceptingInit,
			setVPNIPErr:  someErr,
			wantErr:      true,
			wantState:    state.Failed,
		},
		"no messages sent to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{},
			wantErr:                 true,
			wantState:               state.AcceptingInit,
		},
		"only initial message sent to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{"initialRequest"},
			wantErr:                 true,
			wantState:               state.Failed,
		},
		"wrong initial message sent to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{"stateDiskKey"},
			wantErr:                 true,
			wantState:               state.AcceptingInit,
		},
		"initial message sent twice to node": {
			initialPeers:            []peer.Peer{peer1},
			updatedPeers:            []peer.Peer{peer2},
			state:                   state.AcceptingInit,
			messageSequenceOverride: []string{"initialRequest", "initialRequest"},
			wantErr:                 true,
			wantState:               state.Failed,
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
			fs := afero.NewMemMapFs()
			linuxUserManager := user.NewLinuxUserManagerFake(fs)
			cor := &fakeCore{state: tc.state, vpnPubKey: vpnPubKey, setVPNIPErr: tc.setVPNIPErr, linuxUserManager: linuxUserManager}
			netDialer := testdialer.NewBufconnDialer()
			dialer := grpcutil.NewDialer(fakeValidator{}, netDialer)

			api := New(logger, cor, dialer, nil, nil, nil)
			defer api.Close()

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{peers: tc.updatedPeers, getUpdateErr: tc.getUpdateErr}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(netDialer.GetListener(net.JoinHostPort("10.118.0.1", vpnAPIPort)))
			defer vserver.GracefulStop()

			tlsConfig, err := atls.CreateAttestationServerTLSConfig(&core.MockIssuer{}, nil)
			require.NoError(err)
			pubserver := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
			pubproto.RegisterAPIServer(pubserver, api)
			go pubserver.Serve(netDialer.GetListener(net.JoinHostPort(nodeIP, endpointAVPNPort)))
			defer pubserver.GracefulStop()

			_, nodeVPNPubKey, err := activateNode(require, netDialer, messageSequence, nodeIP, "9000", nodeVPNIP, peer.ToPubProto(tc.initialPeers), ownerID, clusterID, stateDiskKey, ssh.ToProtoSlice(tc.sshKeys))
			assert.Equal(tc.wantState, cor.state)

			if tc.wantErr {
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

			// Test SSH user & key creation. Both cases: "supposed to add" and "not supposed to add"
			// This slightly differs from a real environment (e.g. missing /home) but should be fine in the stub context with a virtual file system
			if tc.sshKeys != nil {
				passwd := user.Passwd{}
				entries, err := passwd.Parse(fs)
				require.NoError(err)
				for _, singleEntry := range entries {
					username := singleEntry.Gecos
					_, err := fs.Stat(fmt.Sprintf("/home/%s/.ssh/authorized_keys.d/ssh-keys", username))
					assert.NoError(err)
				}
			} else {
				passwd := user.Passwd{}
				_, err := passwd.Parse(fs)
				assert.EqualError(err, "open /etc/passwd: file does not exist")
				_, err = fs.Stat("/home")
				assert.EqualError(err, "open /home: file does not exist")
			}
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
		wantErr      bool
	}{
		"basic": {
			peers: peers,
			state: state.IsNode,
		},
		"not activated": {
			peers:   peers,
			state:   state.AcceptingInit,
			wantErr: true,
		},
		"wrong peer kind": {
			peers:   peers,
			state:   state.ActivatingNodes,
			wantErr: true,
		},
		"GetUpdate error": {
			peers:        peers,
			state:        state.IsNode,
			getUpdateErr: someErr,
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state}
			netDialer := testdialer.NewBufconnDialer()
			dialer := grpcutil.NewDialer(fakeValidator{}, netDialer)

			api := New(logger, core, dialer, nil, nil, nil)

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{
				peers:        tc.peers,
				getUpdateErr: tc.getUpdateErr,
			}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(netDialer.GetListener(net.JoinHostPort("10.118.0.1", vpnAPIPort)))
			defer vserver.GracefulStop()

			_, err := api.TriggerNodeUpdate(context.Background(), &pubproto.TriggerNodeUpdateRequest{})
			if tc.wantErr {
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
		wantErr        bool
		wantState      state.State
	}{
		"basic": {
			state:     state.NodeWaitingForClusterJoin,
			wantState: state.IsNode,
		},
		"not activated": {
			state:     state.AcceptingInit,
			wantErr:   true,
			wantState: state.AcceptingInit,
		},
		"wrong peer kind": {
			state:     state.ActivatingNodes,
			wantErr:   true,
			wantState: state.ActivatingNodes,
		},
		"GetK8sJoinArgs error": {
			state:          state.NodeWaitingForClusterJoin,
			getJoinArgsErr: someErr,
			wantErr:        true,
			wantState:      state.NodeWaitingForClusterJoin,
		},
		"JoinCluster error": {
			state:          state.NodeWaitingForClusterJoin,
			joinClusterErr: someErr,
			wantErr:        true,
			wantState:      state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state, joinClusterErr: tc.joinClusterErr}
			netDialer := testdialer.NewBufconnDialer()
			dialer := grpcutil.NewDialer(fakeValidator{}, netDialer)

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
			go vserver.Serve(netDialer.GetListener(net.JoinHostPort("192.0.2.1", vpnAPIPort)))
			defer vserver.GracefulStop()

			_, err := api.JoinCluster(context.Background(), &pubproto.JoinClusterRequest{CoordinatorVpnIp: "192.0.2.1"})

			assert.Equal(tc.wantState, core.state)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal([]kubeadm.BootstrapTokenDiscovery{vapi.joinArgs}, core.joinArgs)
		})
	}
}

func activateNode(require *require.Assertions, dialer netDialer, messageSequence []string, nodeIP, bindPort, nodeVPNIP string, peers []*pubproto.Peer, ownerID, clusterID, stateDiskKey []byte, sshUserKeys []*pubproto.SSHUserKey) (string, []byte, error) {
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
						NodeVpnIp:   nodeVPNIP,
						Peers:       peers,
						OwnerId:     ownerID,
						ClusterId:   clusterID,
						SshUserKeys: sshUserKeys,
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

func dialGRPC(ctx context.Context, dialer netDialer, target string) (*grpc.ClientConn, error) {
	tlsConfig, err := atls.CreateAttestationClientTLSConfig(nil, []atls.Validator{&core.MockValidator{}})
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
	peers            []peer.Peer
	joinArgs         kubeadm.BootstrapTokenDiscovery
	getUpdateErr     error
	getJoinArgsErr   error
	getK8SCertKeyErr error
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

func (a *stubVPNAPI) GetK8SCertificateKey(ctx context.Context, in *vpnproto.GetK8SCertificateKeyRequest) (*vpnproto.GetK8SCertificateKeyResponse, error) {
	return &vpnproto.GetK8SCertificateKeyResponse{CertificateKey: "dummyCertKey"}, a.getK8SCertKeyErr
}

func (a *stubVPNAPI) newServer() *grpc.Server {
	server := grpc.NewServer()
	vpnproto.RegisterAPIServer(server, a)
	return server
}

type netDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
