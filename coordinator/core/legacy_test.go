package core

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/pubapi"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/vpnapi"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

// DEPRECATED test. Don't extend this one, but others or write a new one.
// TODO remove as soon as major changes to this test would be needed.
func TestLegacyActivateCoordinator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	adminVPNKey := []byte{2, 3, 4}

	bufDialer := newBufconnDialer()

	nodeCore1, nodeAPI1, err := newMockCoreWithDialer(bufDialer)
	require.NoError(err)
	defer nodeAPI1.Close()
	_, nodeAPI2, err := newMockCoreWithDialer(bufDialer)
	require.NoError(err)
	defer nodeAPI2.Close()
	_, nodeAPI3, err := newMockCoreWithDialer(bufDialer)
	require.NoError(err)
	defer nodeAPI3.Close()

	nodeIPs := []string{"192.0.2.11", "192.0.2.12", "192.0.2.13"}
	coordinatorIP := "192.0.2.1"
	bindPort := "9000"
	nodeServer1, err := spawnNode(net.JoinHostPort(nodeIPs[0], bindPort), nodeAPI1, bufDialer)
	require.NoError(err)
	defer nodeServer1.GracefulStop()
	nodeServer2, err := spawnNode(net.JoinHostPort(nodeIPs[1], bindPort), nodeAPI2, bufDialer)
	require.NoError(err)
	defer nodeServer2.GracefulStop()
	nodeServer3, err := spawnNode(net.JoinHostPort(nodeIPs[2], bindPort), nodeAPI3, bufDialer)
	require.NoError(err)
	defer nodeServer3.GracefulStop()

	coordinatorCore, coordinatorAPI, err := newMockCoreWithDialer(bufDialer)
	require.NoError(err)
	require.NoError(coordinatorCore.SetVPNIP("10.118.0.1"))
	defer coordinatorAPI.Close()
	coordinatorServer, err := spawnNode(net.JoinHostPort(coordinatorIP, bindPort), coordinatorAPI, bufDialer)
	require.NoError(err)
	defer coordinatorServer.GracefulStop()

	// activate coordinator
	activationReq := &pubproto.ActivateAsCoordinatorRequest{
		AdminVpnPubKey: adminVPNKey,
		NodePublicIps:  nodeIPs,
		MasterSecret:   []byte("Constellation"),
		KmsUri:         kms.ClusterKMSURI,
		StorageUri:     kms.NoStoreURI,
	}
	testActivationSvr := &stubAVPNActivateCoordinatorServer{}
	assert.NoError(coordinatorAPI.ActivateAsCoordinator(activationReq, testActivationSvr))

	// Coordinator sets own key
	coordinatorKey, err := coordinatorCore.data().GetVPNKey()
	assert.NoError(err)

	// Coordinator streams admin conf
	require.NotEmpty(testActivationSvr.sent)
	adminConfig := testActivationSvr.sent[len(testActivationSvr.sent)-1].GetAdminConfig()
	require.NotNil(adminConfig)
	assert.NotEmpty(adminConfig.AdminVpnIp)
	assert.Equal(coordinatorKey, adminConfig.CoordinatorVpnPubKey)
	assert.NotNil(adminConfig.Kubeconfig)
	require.NotNil(testActivationSvr.sent[0])
	require.NotNil(testActivationSvr.sent[0].GetLog())
	assert.NotEmpty(testActivationSvr.sent[0].GetLog().Message)

	// Coordinator cannot be activated a second time
	assert.Error(coordinatorAPI.ActivateAsCoordinator(activationReq, testActivationSvr))

	// Node cannot be activated a second time
	nodeResp, err := nodeAPI3.ActivateAsNode(context.TODO(), &pubproto.ActivateAsNodeRequest{
		NodeVpnIp: "192.0.2.1:9004",
		Peers: []*pubproto.Peer{{
			VpnPubKey: coordinatorKey,
			PublicIp:  coordinatorIP,
			VpnIp:     "10.118.0.1",
		}},
		OwnerId:   []byte("ownerID"),
		ClusterId: []byte("clusterID"),
	})
	assert.Error(err)
	assert.Nil(nodeResp)

	// Assert Coordinator
	peers := coordinatorCore.vpn.(*stubVPN).peers
	assert.Less(3, len(peers))
	// coordinator peers contain admin
	found := false
	for _, peer := range peers {
		if bytes.Equal(adminVPNKey, peer.pubKey) {
			found = true
			break
		}
	}
	assert.True(found)

	// Assert Node
	peers = nodeCore1.vpn.(*stubVPN).peers
	assert.Less(0, len(peers))
	assert.NotEmpty(peers[0].publicIP)
	assert.Equal(coordinatorKey, peers[0].pubKey)
}

// newMockCoreWithDialer creates a new core object with attestation mock and provided dialer for testing.
func newMockCoreWithDialer(dialer *bufconnDialer) (*Core, *pubapi.API, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, nil, err
	}

	validator := NewMockValidator()
	vpn := &stubVPN{}
	kubeFake := &ClusterFake{}
	metadataFake := &ProviderMetadataFake{}
	ccmFake := &CloudControllerManagerFake{}
	cnmFake := &CloudNodeManagerFake{}
	autoscalerFake := &ClusterAutoscalerFake{}
	encryptedDiskFake := &EncryptedDiskFake{}

	getPublicAddr := func() (string, error) {
		return "192.0.2.1", nil
	}
	core, err := NewCore(vpn, kubeFake, metadataFake, ccmFake, cnmFake, autoscalerFake, encryptedDiskFake, zapLogger, vtpm.OpenSimulatedTPM, &fakeStoreFactory{}, file.NewHandler(afero.NewMemMapFs()))
	if err != nil {
		return nil, nil, err
	}
	if err := core.AdvanceState(state.AcceptingInit, nil, nil); err != nil {
		return nil, nil, err
	}

	vapiServer := &fakeVPNAPIServer{logger: zapLogger, core: core, dialer: dialer}
	papi := pubapi.New(zapLogger, core, dialer, vapiServer, validator, getPublicAddr)

	return core, papi, nil
}

type bufconnDialer struct {
	mut       sync.Mutex
	listeners map[string]*bufconn.Listener
}

func newBufconnDialer() *bufconnDialer {
	return &bufconnDialer{listeners: make(map[string]*bufconn.Listener)}
}

func (b *bufconnDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	b.mut.Lock()
	listener, ok := b.listeners[address]
	b.mut.Unlock()
	if !ok {
		return nil, fmt.Errorf("could not connect to server on %v", address)
	}
	return listener.DialContext(ctx)
}

func (b *bufconnDialer) addListener(endpoint string, listener *bufconn.Listener) {
	b.mut.Lock()
	b.listeners[endpoint] = listener
	b.mut.Unlock()
}

func spawnNode(endpoint string, testNodeCore *pubapi.API, bufDialer *bufconnDialer) (*grpc.Server, error) {
	tlsConfig, err := atls.CreateAttestationServerTLSConfig(&MockIssuer{})
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pubproto.RegisterAPIServer(grpcServer, testNodeCore)

	const bufferSize = 8 * 1024
	listener := bufconn.Listen(bufferSize)
	bufDialer.addListener(endpoint, listener)

	log.Printf("bufconn server listening at %v", endpoint)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return grpcServer, nil
}

type stubAVPNActivateCoordinatorServer struct {
	grpc.ServerStream

	sendErr error

	sent []*pubproto.ActivateAsCoordinatorResponse
}

func (s *stubAVPNActivateCoordinatorServer) Send(req *pubproto.ActivateAsCoordinatorResponse) error {
	s.sent = append(s.sent, req)
	return s.sendErr
}

type fakeVPNAPIServer struct {
	logger   *zap.Logger
	core     vpnapi.Core
	dialer   *bufconnDialer
	listener net.Listener
	server   *grpc.Server
}

func (v *fakeVPNAPIServer) Listen(endpoint string) error {
	api := vpnapi.New(v.logger, v.core)
	v.server = grpc.NewServer()
	vpnproto.RegisterAPIServer(v.server, api)
	listener := bufconn.Listen(1024)
	v.dialer.addListener(endpoint, listener)
	v.listener = listener
	return nil
}

func (v *fakeVPNAPIServer) Serve() error {
	return v.server.Serve(v.listener)
}

func (v *fakeVPNAPIServer) Close() {
	if v.server != nil {
		v.server.GracefulStop()
	}
}
