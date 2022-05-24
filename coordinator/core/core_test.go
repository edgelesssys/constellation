package core

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/attestation/simulator"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/nodestate"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/edgelesssys/constellation/coordinator/util/grpcutil"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/file"
	kms "github.com/edgelesssys/constellation/kms/server/setup"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestGetNextNodeIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.NewMemMapFs()
	core, err := NewCore(&stubVPN{}, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
	require.NoError(err)
	require.NoError(core.InitializeStoreIPs())

	ip, err := core.GetNextNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.11", ip)

	ip, err = core.GetNextNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.12", ip)

	ip, err = core.GetNextNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.13", ip)

	require.NoError(core.data().PutFreedNodeVPNIP("10.118.0.12"))
	require.NoError(core.data().PutFreedNodeVPNIP("10.118.0.13"))
	ipsInStore := map[string]struct{}{
		"10.118.0.13": {},
		"10.118.0.12": {},
	}

	ip, err = core.GetNextNodeIP()
	assert.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = core.GetNextNodeIP()
	assert.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = core.GetNextNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.14", ip)
}

func TestSwitchToPersistentStore(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	storeFactory := &fakeStoreFactory{}
	fs := afero.NewMemMapFs()
	core, err := NewCore(&stubVPN{}, nil, nil, nil, zaptest.NewLogger(t), nil, storeFactory, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
	require.NoError(core.store.Put("test", []byte("test")))
	require.NoError(err)

	require.NoError(core.SwitchToPersistentStore())
	value, err := core.store.Get("test")
	assert.NoError(err)
	assert.Equal("test", string(value))
}

func TestGetIDs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.NewMemMapFs()
	core, err := NewCore(&stubVPN{}, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
	require.NoError(err)

	_, _, err = core.GetIDs(nil)
	assert.Error(err)

	masterSecret := []byte{2, 3, 4}
	ownerID1, clusterID1, err := core.GetIDs(masterSecret)
	require.NoError(err)
	require.NotEmpty(ownerID1)
	require.NotEmpty(clusterID1)

	require.NoError(core.data().PutClusterID(clusterID1))

	ownerID2, clusterID2, err := core.GetIDs(nil)
	require.NoError(err)
	assert.Equal(ownerID1, ownerID2)
	assert.Equal(clusterID1, clusterID2)
}

func TestNotifyNodeHeartbeat(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.NewMemMapFs()
	core, err := NewCore(&stubVPN{}, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
	require.NoError(err)

	const ip = "192.0.2.1"
	assert.Empty(core.lastHeartbeats)
	core.NotifyNodeHeartbeat(&net.IPAddr{IP: net.ParseIP(ip)})
	assert.Contains(core.lastHeartbeats, ip)
}

func TestDeriveKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.NewMemMapFs()
	core, err := NewCore(&stubVPN{}, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
	require.NoError(err)

	// error when no kms is set up
	_, err = core.GetDataKey(context.Background(), "key-1", 32)
	assert.Error(err)

	kms := &fakeKMS{}
	core.kms = kms

	require.NoError(core.store.Put("kekID", []byte("master-key")))

	// error when no master secret is set
	_, err = core.GetDataKey(context.Background(), "key-1", 32)
	assert.Error(err)
	err = core.kms.CreateKEK(context.Background(), "master-key", []byte("Constellation"))
	require.NoError(err)

	key, err := core.GetDataKey(context.Background(), "key-1", 32)
	assert.NoError(err)
	assert.Equal(kms.dek, key)

	kms.getDEKErr = errors.New("error")
	_, err = core.GetDataKey(context.Background(), "key-1", 32)
	assert.Error(err)
}

func TestInitialize(t *testing.T) {
	testCases := map[string]struct {
		initializePCRs bool
		writeNodeState bool
		role           role.Role
		wantActivated  bool
		wantState      state.State
		wantErr        bool
	}{
		"fresh node": {
			wantState: state.AcceptingInit,
		},
		"activated coordinator": {
			initializePCRs: true,
			writeNodeState: true,
			role:           role.Coordinator,
			wantActivated:  true,
			wantState:      state.ActivatingNodes,
		},
		"activated node": {
			initializePCRs: true,
			writeNodeState: true,
			role:           role.Node,
			wantActivated:  true,
			wantState:      state.IsNode,
		},
		"activated node with no node state": {
			initializePCRs: true,
			writeNodeState: false,
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			openTPM, simulatedTPMCloser := simulator.NewSimulatedTPMOpenFunc()
			defer simulatedTPMCloser.Close()
			if tc.initializePCRs {
				require.NoError(vtpm.MarkNodeAsInitialized(openTPM, []byte{0x0, 0x1, 0x2, 0x3}, []byte{0x4, 0x5, 0x6, 0x7}))
			}
			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			if tc.writeNodeState {
				require.NoError((&nodestate.NodeState{
					Role:       tc.role,
					VPNPrivKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7},
				}).ToFile(fileHandler))
			}
			core, err := NewCore(&stubVPN{}, &clusterStub{}, &ProviderMetadataFake{}, nil, zaptest.NewLogger(t), openTPM, &fakeStoreFactory{}, fileHandler, user.NewLinuxUserManagerFake(fs))
			require.NoError(err)
			core.initialVPNPeersRetriever = fakeInitializeVPNPeersRetriever
			// prepare store to emulate initialized KMS
			require.NoError(core.data().PutKMSData(kms.KMSInformation{StorageUri: kms.NoStoreURI, KmsUri: kms.ClusterKMSURI}))
			require.NoError(core.data().PutMasterSecret([]byte("master-secret")))
			dialer := grpcutil.NewDialer(&MockValidator{}, testdialer.NewBufconnDialer())

			nodeActivated, err := core.Initialize(context.Background(), dialer, &stubPubAPI{})
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantActivated, nodeActivated)
			assert.Equal(tc.wantState, core.state)
		})
	}
}

func TestPersistNodeState(t *testing.T) {
	testCases := map[string]struct {
		vpn            VPN
		touchStateFile bool
		wantErr        bool
	}{
		"persisting works": {
			vpn: &stubVPN{
				privateKey: []byte("private-key"),
			},
		},
		"retrieving VPN key fails": {
			vpn: &stubVPN{
				getPrivateKeyErr: errors.New("error"),
			},
			wantErr: true,
		},
		"writing node state over existing file fails": {
			vpn: &stubVPN{
				privateKey: []byte("private-key"),
			},
			touchStateFile: true,
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			if tc.touchStateFile {
				file, err := fs.Create("/run/state/constellation/node_state.json")
				require.NoError(err)
				require.NoError(file.Close())
			}
			core, err := NewCore(tc.vpn, nil, nil, nil, zaptest.NewLogger(t), nil, nil, fileHandler, user.NewLinuxUserManagerFake(fs))
			require.NoError(err)
			err = core.PersistNodeState(role.Coordinator, "192.0.2.1", []byte("owner-id"), []byte("cluster-id"))
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			nodeState, err := nodestate.FromFile(fileHandler)
			assert.NoError(err)
			assert.Equal(nodestate.NodeState{
				Role:       role.Coordinator,
				VPNIP:      "192.0.2.1",
				VPNPrivKey: []byte("private-key"),
				OwnerID:    []byte("owner-id"),
				ClusterID:  []byte("cluster-id"),
			}, *nodeState)
		})
	}
}

type fakeStoreFactory struct {
	store store.Store
}

func (f *fakeStoreFactory) New() (store.Store, error) {
	f.store = store.NewStdStore()
	return f.store, nil
}

type fakeKMS struct {
	kek       []byte
	dek       []byte
	getDEKErr error
}

func (k *fakeKMS) CreateKEK(ctx context.Context, keyID string, key []byte) error {
	k.kek = []byte(keyID)
	return nil
}

func (k *fakeKMS) GetDEK(ctx context.Context, kekID, keyID string, length int) ([]byte, error) {
	if k.getDEKErr != nil {
		return nil, k.getDEKErr
	}
	if k.kek == nil {
		return nil, errors.New("error")
	}
	return k.dek, nil
}

type stubPubAPI struct {
	startVPNAPIErr error
}

func (p *stubPubAPI) StartVPNAPIServer(vpnIP string) error {
	return p.startVPNAPIErr
}

func (p *stubPubAPI) StartUpdateLoop() {}

func fakeInitializeVPNPeersRetriever(ctx context.Context, dialer Dialer, logger *zap.Logger, metadata ProviderMetadata, ownCoordinatorEndpoint *string) ([]peer.Peer, error) {
	return []peer.Peer{}, nil
}
