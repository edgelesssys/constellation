package core

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/nodestate"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/edgelesssys/constellation/coordinator/storewrapper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap/zaptest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		// https://github.com/kubernetes/klog/issues/282, https://github.com/kubernetes/klog/issues/188
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

func TestAddAdmin(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	vpn := &stubVPN{}
	core, err := NewCore(vpn, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
	require.NoError(err)
	require.NoError(core.InitializeStoreIPs())

	pubKey := []byte{2, 3, 4}

	vpnIP, err := core.AddAdmin(pubKey)
	require.NoError(err)
	assert.NotNil(net.ParseIP(vpnIP))
	assert.Equal([]stubVPNPeer{{pubKey: pubKey, vpnIP: vpnIP}}, vpn.peers)
}

func TestGetNextNodeIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
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
	core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, storeFactory, file.NewHandler(afero.NewMemMapFs()))
	require.NoError(err)

	require.NoError(core.SwitchToPersistentStore())

	key, err := storewrapper.StoreWrapper{Store: storeFactory.store}.GetVPNKey()
	require.NoError(err)
	assert.NotEmpty(key)
}

func TestGetIDs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
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

	core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
	require.NoError(err)

	const ip = "192.0.2.1"
	assert.Empty(core.lastHeartbeats)
	core.NotifyNodeHeartbeat(&net.IPAddr{IP: net.ParseIP(ip)})
	assert.Contains(core.lastHeartbeats, ip)
}

func TestDeriveKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
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
		initializePCRs  bool
		writeNodeState  bool
		role            role.Role
		expectActivated bool
		expectedState   state.State
		expectPanic     bool
		expectErr       bool
	}{
		"fresh node": {
			expectedState: state.AcceptingInit,
		},
		"activated coordinator": {
			initializePCRs:  true,
			writeNodeState:  true,
			role:            role.Coordinator,
			expectPanic:     true, // TODO: adapt test case once restart is implemented
			expectActivated: true,
			expectedState:   state.ActivatingNodes,
		},
		"activated node": {
			initializePCRs:  true,
			writeNodeState:  true,
			role:            role.Node,
			expectPanic:     true, // TODO: adapt test case once restart is implemented
			expectActivated: true,
			expectedState:   state.IsNode,
		},
		"activated node with no node state": {
			initializePCRs: true,
			writeNodeState: false,
			expectErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			openTPM, simulatedTPMCloser := vtpm.NewSimulatedTPMOpenFunc()
			defer simulatedTPMCloser.Close()
			if tc.initializePCRs {
				require.NoError(vtpm.MarkNodeAsInitialized(openTPM, []byte{0x0, 0x1, 0x2, 0x3}, []byte{0x4, 0x5, 0x6, 0x7}))
			}
			fileHandler := file.NewHandler(afero.NewMemMapFs())
			if tc.writeNodeState {
				require.NoError((&nodestate.NodeState{
					Role:       tc.role,
					VPNPrivKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7},
				}).ToFile(fileHandler))
			}

			core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, zaptest.NewLogger(t), openTPM, nil, fileHandler)
			require.NoError(err)

			if tc.expectPanic {
				assert.Panics(func() { _, _ = core.Initialize() })
				return
			}

			nodeActivated, err := core.Initialize()
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectActivated, nodeActivated)
			assert.Equal(tc.expectedState, core.state)
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
