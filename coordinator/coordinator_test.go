package coordinator

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/simulator"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/edgelesssys/constellation/coordinator/vpnapi"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		// https://github.com/kubernetes/klog/issues/282, https://github.com/kubernetes/klog/issues/188
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

// TestCoordinator tests the integration of packages core, pubapi, and vpnapi. It activates
// a coordinator and some nodes and (virtually) sends a packet over the fake VPN.
func TestCoordinator(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	nodeIPs := []string{"192.0.2.11", "192.0.2.12", "192.0.2.13"}
	coordinatorIP := "192.0.2.1"
	bindPort := "9000"
	logger := zaptest.NewLogger(t)
	dialer := testdialer.NewBufconnDialer()
	netw := newNetwork()

	// spawn 4 peers: 1 designated coordinator and 3 nodes
	coordServer, coordPAPI, _ := spawnPeer(require, logger.Named("coord"), dialer, netw, net.JoinHostPort(coordinatorIP, bindPort))
	defer coordPAPI.Close()
	defer coordServer.GracefulStop()
	nodeServer1, nodePAPI1, nodeVPN1 := spawnPeer(require, logger.Named("node1"), dialer, netw, net.JoinHostPort(nodeIPs[0], bindPort))
	defer nodePAPI1.Close()
	defer nodeServer1.GracefulStop()
	nodeServer2, nodePAPI2, nodeVPN2 := spawnPeer(require, logger.Named("node2"), dialer, netw, net.JoinHostPort(nodeIPs[1], bindPort))
	defer nodePAPI2.Close()
	defer nodeServer2.GracefulStop()
	nodeServer3, nodePAPI3, nodeVPN3 := spawnPeer(require, logger.Named("node3"), dialer, netw, net.JoinHostPort(nodeIPs[2], bindPort))
	defer nodePAPI3.Close()
	defer nodeServer3.GracefulStop()

	require.NoError(activateCoordinator(require, dialer, coordinatorIP, bindPort, nodeIPs))

	// send something from node 1 to node 2

	nodeIP1, err := nodeVPN1.GetInterfaceIP()
	require.NoError(err)
	nodeIP2, err := nodeVPN2.GetInterfaceIP()
	require.NoError(err)
	assert.NotEqual(nodeIP1, nodeIP2)

	nodeVPN1.send(nodeIP2, "foo")
	assert.Nil(nodeVPN3.recv())
	pa := nodeVPN2.recv()
	require.NotNil(pa)
	assert.Equal(nodeIP1, pa.src)
	assert.Equal("foo", pa.data)
}

// TestConcurrent is supposed to detect data races when run with -race.
func TestConcurrent(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	nodeIPs := []string{"192.0.2.11", "192.0.2.12", "192.0.2.13"}
	coordinatorIP := "192.0.2.1"
	bindPort := "9000"
	logger := zaptest.NewLogger(t)
	dialer := testdialer.NewBufconnDialer()
	netw := newNetwork()

	// spawn peers
	coordServer, coordPAPI, _ := spawnPeer(require, logger.Named("coord"), dialer, netw, net.JoinHostPort(coordinatorIP, bindPort))
	defer coordPAPI.Close()
	defer coordServer.GracefulStop()
	nodeServer1, nodePAPI1, _ := spawnPeer(require, logger.Named("node1"), dialer, netw, net.JoinHostPort(nodeIPs[0], bindPort))
	defer nodePAPI1.Close()
	defer nodeServer1.GracefulStop()
	nodeServer2, nodePAPI2, _ := spawnPeer(require, logger.Named("node2"), dialer, netw, net.JoinHostPort(nodeIPs[1], bindPort))
	defer nodePAPI2.Close()
	defer nodeServer2.GracefulStop()

	var wg sync.WaitGroup

	actCoord := func() {
		defer wg.Done()
		_ = activateCoordinator(require, dialer, coordinatorIP, bindPort, nodeIPs)
	}

	actNode := func(target string) {
		defer wg.Done()
		conn, _ := dialGRPC(context.Background(), dialer, target)
		defer conn.Close()
		client := pubproto.NewAPIClient(conn)
		stream, err := client.ActivateAsNode(context.Background())
		assert.NoError(err)
		assert.NoError(stream.Send(&pubproto.ActivateAsNodeRequest{
			Request: &pubproto.ActivateAsNodeRequest_InitialRequest{},
		}))
		_, err = stream.Recv()
		assert.Error(err)
	}

	updNode := func(papi *pubapi.API, noerr bool) {
		defer wg.Done()
		_, err := papi.TriggerNodeUpdate(context.Background(), &pubproto.TriggerNodeUpdateRequest{})
		if noerr {
			assert.NoError(err)
		}
	}

	getState := func(papi *pubapi.API) {
		defer wg.Done()
		_, err := papi.GetState(context.Background(), &pubproto.GetStateRequest{})
		assert.NoError(err)
	}

	join := func(papi *pubapi.API) {
		defer wg.Done()
		_, err := papi.JoinCluster(context.Background(), &pubproto.JoinClusterRequest{})
		assert.Error(err)
	}

	// activate coordinator and make some other calls concurrently
	wg.Add(16)
	go actCoord()
	go actCoord()
	go updNode(coordPAPI, false)
	go updNode(coordPAPI, false)
	go updNode(nodePAPI1, false)
	go updNode(nodePAPI1, false)
	go updNode(nodePAPI2, false)
	go updNode(nodePAPI2, false)
	go getState(coordPAPI)
	go getState(coordPAPI)
	go getState(nodePAPI1)
	go getState(nodePAPI1)
	go getState(nodePAPI2)
	go getState(nodePAPI2)
	go join(coordPAPI)
	go join(coordPAPI)
	wg.Wait()

	// make some concurrent calls on the activated peers
	wg.Add(26)
	go actCoord()
	go actCoord()
	go actNode(net.JoinHostPort(coordinatorIP, bindPort))
	go actNode(net.JoinHostPort(coordinatorIP, bindPort))
	go actNode(net.JoinHostPort(nodeIPs[0], bindPort))
	go actNode(net.JoinHostPort(nodeIPs[0], bindPort))
	go actNode(net.JoinHostPort(nodeIPs[1], bindPort))
	go actNode(net.JoinHostPort(nodeIPs[1], bindPort))
	go updNode(coordPAPI, false)
	go updNode(coordPAPI, false)
	go updNode(nodePAPI1, true)
	go updNode(nodePAPI1, true)
	go updNode(nodePAPI2, true)
	go updNode(nodePAPI2, true)
	go getState(coordPAPI)
	go getState(coordPAPI)
	go getState(nodePAPI1)
	go getState(nodePAPI1)
	go getState(nodePAPI2)
	go getState(nodePAPI2)
	go join(coordPAPI)
	go join(coordPAPI)
	go join(nodePAPI1)
	go join(nodePAPI1)
	go join(nodePAPI2)
	go join(nodePAPI2)
	wg.Wait()
}

func spawnPeer(require *require.Assertions, logger *zap.Logger, dialer *testdialer.BufconnDialer, netw *network, endpoint string) (*grpc.Server, *pubapi.API, *fakeVPN) {
	vpn := newVPN(netw, endpoint)
	cor, err := core.NewCore(vpn, &core.ClusterFake{}, &core.ProviderMetadataFake{}, &core.CloudControllerManagerFake{}, &core.CloudNodeManagerFake{}, &core.ClusterAutoscalerFake{}, &core.EncryptedDiskFake{}, logger, simulator.OpenSimulatedTPM, fakeStoreFactory{}, file.NewHandler(afero.NewMemMapFs()))
	require.NoError(err)
	require.NoError(cor.AdvanceState(state.AcceptingInit, nil, nil))

	getPublicAddr := func() (string, error) {
		return "192.0.2.1", nil
	}

	vapiServer := &fakeVPNAPIServer{logger: logger.Named("vpnapi"), core: cor, dialer: dialer}
	papi := pubapi.New(logger, cor, dialer, vapiServer, &core.MockValidator{}, getPublicAddr, nil)

	tlsConfig, err := atls.CreateAttestationServerTLSConfig(&core.MockIssuer{})
	require.NoError(err)
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pubproto.RegisterAPIServer(server, papi)

	listener := dialer.GetListener(endpoint)
	go server.Serve(listener)

	return server, papi, vpn
}

func activateCoordinator(require *require.Assertions, dialer pubapi.Dialer, coordinatorIP, bindPort string, nodeIPs []string) error {
	ctx := context.Background()
	conn, err := dialGRPC(ctx, dialer, net.JoinHostPort(coordinatorIP, bindPort))
	require.NoError(err)
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	stream, err := client.ActivateAsCoordinator(ctx, &pubproto.ActivateAsCoordinatorRequest{
		NodePublicIps: nodeIPs,
		MasterSecret:  []byte("Constellation"),
		KmsUri:        kms.ClusterKMSURI,
		StorageUri:    kms.NoStoreURI,
	})
	require.NoError(err)

	for {
		_, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func dialGRPC(ctx context.Context, dialer pubapi.Dialer, target string) (*grpc.ClientConn, error) {
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

type fakeStoreFactory struct{}

func (fakeStoreFactory) New() (store.Store, error) {
	return store.NewStdStore(), nil
}

type fakeVPNAPIServer struct {
	logger   *zap.Logger
	core     vpnapi.Core
	dialer   *testdialer.BufconnDialer
	listener net.Listener
	server   *grpc.Server
}

func (v *fakeVPNAPIServer) Listen(endpoint string) error {
	api := vpnapi.New(v.logger, v.core)
	v.server = grpc.NewServer()
	vpnproto.RegisterAPIServer(v.server, api)
	v.listener = v.dialer.GetListener(endpoint)
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

type network struct {
	packets map[string][]packet
}

func newNetwork() *network {
	return &network{packets: make(map[string][]packet)}
}

type packet struct {
	src  string
	data string
}

type fakeVPN struct {
	peers       map[string]string // vpnIP -> publicIP
	netw        *network
	publicIP    string
	interfaceIP string
}

func newVPN(netw *network, publicEndpoint string) *fakeVPN {
	publicIP, _, err := net.SplitHostPort(publicEndpoint)
	if err != nil {
		panic(err)
	}
	return &fakeVPN{
		peers:    make(map[string]string),
		netw:     netw,
		publicIP: publicIP,
	}
}

func (*fakeVPN) Setup(privKey []byte) error {
	return nil
}

func (*fakeVPN) GetPrivateKey() ([]byte, error) {
	return nil, nil
}

func (*fakeVPN) GetPublicKey() ([]byte, error) {
	return nil, nil
}

func (v *fakeVPN) GetInterfaceIP() (string, error) {
	return v.interfaceIP, nil
}

func (v *fakeVPN) SetInterfaceIP(ip string) error {
	v.interfaceIP = ip
	return nil
}

func (v *fakeVPN) AddPeer(pubKey []byte, publicIP string, vpnIP string) error {
	v.peers[vpnIP] = publicIP
	return nil
}

func (v *fakeVPN) RemovePeer(pubKey []byte) error {
	panic("dummy")
}

func (v *fakeVPN) UpdatePeers(peers []peer.Peer) error {
	for _, peer := range peers {
		if err := v.AddPeer(peer.VPNPubKey, peer.PublicIP, peer.VPNIP); err != nil {
			return err
		}
	}
	return nil
}

func (v *fakeVPN) send(dst string, data string) {
	pubdst := v.peers[dst]
	packets := v.netw.packets
	packets[pubdst] = append(packets[pubdst], packet{src: v.publicIP, data: data})
}

func (v *fakeVPN) recv() *packet {
	packets := v.netw.packets
	queue := packets[v.publicIP]
	if len(queue) == 0 {
		return nil
	}
	packet := queue[0]
	packets[v.publicIP] = queue[1:]
	for vpnIP, pubIP := range v.peers {
		if pubIP == packet.src {
			packet.src = vpnIP
		}
	}
	return &packet
}
