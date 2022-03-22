package storewrapper

import (
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/kubernetes/klog/issues/282, https://github.com/kubernetes/klog/issues/188
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

func TestStoreWrapper(t *testing.T) {
	assert := assert.New(t)

	curState := state.IsNode

	key := []byte{2, 3, 4}
	masterSecret := []byte("Constellation")

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))
	dummyKey, err := wgtypes.GenerateKey()
	assert.NoError(err)
	assert.NoError(stwrapper.PutVPNKey(dummyKey[:]))
	assert.NoError(stwrapper.PutMasterSecret(masterSecret))

	// save values to store
	tx, err := stor.BeginTransaction()
	assert.NoError(err)
	txdata := StoreWrapper{tx}
	assert.NoError(txdata.PutState(curState))
	assert.NoError(txdata.PutVPNKey(key))
	assert.NoError(tx.Commit())

	// see if we can retrieve them again
	savedState, err := stwrapper.GetState()
	assert.NoError(err)
	assert.Equal(curState, savedState)
	savedKey, err := stwrapper.GetVPNKey()
	assert.NoError(err)
	assert.Equal(key, savedKey)
	savedSecret, err := stwrapper.GetMasterSecret()
	assert.NoError(err)
	assert.Equal(masterSecret, savedSecret)
}

func TestStoreWrapperDefaults(t *testing.T) {
	assert := assert.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))
	dummyKey, err := wgtypes.GenerateKey()
	assert.NoError(err)
	assert.NoError(stwrapper.PutVPNKey(dummyKey[:]))

	statevalue, err := stwrapper.GetState()
	assert.NoError(err)
	assert.Equal(state.AcceptingInit, statevalue)

	k, err := stwrapper.GetVPNKey()
	assert.NoError(err)
	assert.NotEmpty(k)

	// Nothing else was set, should always return error
}

func TestStoreWrapperRollback(t *testing.T) {
	assert := assert.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))
	dummyKey, err := wgtypes.GenerateKey()
	assert.NoError(err)
	assert.NoError(stwrapper.PutVPNKey(dummyKey[:]))

	k1 := []byte{2, 3, 4}
	k2 := []byte{3, 4, 5}

	tx, err := stor.BeginTransaction()
	assert.NoError(err)
	assert.NoError(StoreWrapper{tx}.PutVPNKey(k1))
	assert.NoError(tx.Commit())

	tx, err = stor.BeginTransaction()
	assert.NoError(err)
	assert.NoError(StoreWrapper{tx}.PutVPNKey(k2))
	tx.Rollback()

	val, err := stwrapper.GetVPNKey()
	assert.NoError(err)
	assert.Equal(k1, val)
}

func TestStoreWrapperPeerInterface(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))
	dummyKey, err := wgtypes.GenerateKey()
	assert.NoError(err)
	assert.NoError(stwrapper.PutVPNKey(dummyKey[:]))

	key, err := wgtypes.GeneratePrivateKey()
	assert.NoError(err)

	ip := "192.0.2.1"
	internalIP := "10.118.2.0"

	validPeer := peer.Peer{
		PublicEndpoint: ip,
		VPNPubKey:      key[:],
		VPNIP:          internalIP,
	}
	require.NoError(stwrapper.PutPeer(validPeer))
	data, err := stwrapper.GetPeers()
	require.NoError(err)
	require.Equal(1, len(data))
	assert.Equal(ip, data[0].PublicEndpoint)
	assert.Equal(key[:], data[0].VPNPubKey)
	assert.Equal(internalIP, data[0].VPNIP)

	invalidPeer := peer.Peer{
		PublicEndpoint: ip,
		VPNPubKey:      key[:],
		VPNIP:          "",
	}
	assert.Error(stwrapper.PutPeer(invalidPeer))
}

func TestStoreWrapperGetVPNIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutFreedNodeVPNIP("203.0.113.1"))
	require.NoError(stwrapper.PutFreedNodeVPNIP("203.0.113.2"))
	ipsInStore := map[string]struct{}{
		"203.0.113.1": {},
		"203.0.113.2": {},
	}

	ip, err := stwrapper.GetFreedNodeVPNIP()
	require.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.GetFreedNodeVPNIP()
	require.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.GetFreedNodeVPNIP()
	assert.NoError(err)
	assert.Len(ip, 0)
}

func TestGenerateNextIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutLastNodeIP([]byte{10, 118, 0, 1}))

	ip, err := stwrapper.generateNextNodeIP()
	assert.NoError(err)
	assert.Equal(ip, "10.118.0.2")

	ip, err = stwrapper.generateNextNodeIP()
	assert.NoError(err)
	assert.Equal(ip, "10.118.0.3")

	for i := 0; i < 256*256-7; i++ {
		ip, err = stwrapper.generateNextNodeIP()
		assert.NoError(err)
		assert.NotEmpty(ip)
	}

	ip, err = stwrapper.generateNextNodeIP()
	assert.NoError(err)
	assert.Equal(ip, "10.118.255.253")

	ip, err = stwrapper.generateNextNodeIP()
	assert.NoError(err)
	assert.Equal(ip, "10.118.255.254")

	// 10.118.255.255 (broadcast IP) should not be returned
	ip, err = stwrapper.generateNextNodeIP()
	assert.Error(err)
	assert.Empty(ip)

	// error should still persist
	ip, err = stwrapper.generateNextNodeIP()
	assert.Error(err)
	assert.Empty(ip)
}

func TestPopNextFreeIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutLastNodeIP([]byte{10, 118, 0, 1}))

	ip, err := stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.2", ip)

	ip, err = stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.3", ip)

	require.NoError(stwrapper.PutFreedNodeVPNIP("10.118.0.3"))
	require.NoError(stwrapper.PutFreedNodeVPNIP("10.118.0.2"))
	ipsInStore := map[string]struct{}{
		"10.118.0.3": {},
		"10.118.0.2": {},
	}

	ip, err = stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Equal("10.118.0.4", ip)
}
