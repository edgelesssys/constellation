package storewrapper

import (
	"net/netip"
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
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestStoreWrapper(t *testing.T) {
	assert := assert.New(t)

	curState := state.IsNode

	masterSecret := []byte("Constellation")

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))
	assert.NoError(stwrapper.PutMasterSecret(masterSecret))

	// save values to store
	tx, err := stor.BeginTransaction()
	assert.NoError(err)
	txdata := StoreWrapper{tx}
	assert.NoError(txdata.PutState(curState))
	assert.NoError(tx.Commit())

	// see if we can retrieve them again
	savedState, err := stwrapper.GetState()
	assert.NoError(err)
	assert.Equal(curState, savedState)
	savedSecret, err := stwrapper.GetMasterSecret()
	assert.NoError(err)
	assert.Equal(masterSecret, savedSecret)
}

func TestStoreWrapperDefaults(t *testing.T) {
	assert := assert.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))

	statevalue, err := stwrapper.GetState()
	assert.NoError(err)
	assert.Equal(state.AcceptingInit, statevalue)
}

func TestStoreWrapperRollback(t *testing.T) {
	assert := assert.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))

	assert.NoError(stwrapper.PutClusterID([]byte{1, 2, 3}))

	c1 := []byte{2, 3, 4}
	c2 := []byte{3, 4, 5}

	tx, err := stor.BeginTransaction()
	assert.NoError(err)
	assert.NoError(StoreWrapper{tx}.PutClusterID(c1))
	assert.NoError(tx.Commit())

	tx, err = stor.BeginTransaction()
	assert.NoError(err)
	assert.NoError(StoreWrapper{tx}.PutClusterID(c2))
	tx.Rollback()

	val, err := stwrapper.GetClusterID()
	assert.NoError(err)
	assert.Equal(c1, val)
}

func TestStoreWrapperPeerInterface(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	assert.NoError(stwrapper.PutState(state.AcceptingInit))

	key, err := wgtypes.GeneratePrivateKey()
	assert.NoError(err)

	ip := "192.0.2.1"
	internalIP := "10.118.2.0"

	validPeer := peer.Peer{
		PublicIP:  ip,
		VPNPubKey: key[:],
		VPNIP:     internalIP,
	}
	require.NoError(stwrapper.PutPeer(validPeer))
	data, err := stwrapper.GetPeers()
	require.NoError(err)
	require.Equal(1, len(data))
	assert.Equal(ip, data[0].PublicIP)
	assert.Equal(key[:], data[0].VPNPubKey)
	assert.Equal(internalIP, data[0].VPNIP)

	invalidPeer := peer.Peer{
		PublicIP:  ip,
		VPNPubKey: key[:],
		VPNIP:     "",
	}
	assert.Error(stwrapper.PutPeer(invalidPeer))
}

func TestGenerateNextNodeIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutNextNodeIP(netip.AddrFrom4([4]byte{10, 118, 0, 11})))

	ip, err := stwrapper.getNextNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 11}), ip)

	ip, err = stwrapper.getNextNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 12}), ip)

	ip, err = stwrapper.getNextNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 13}), ip)

	for i := 0; i < 256*256-17; i++ {
		ip, err = stwrapper.getNextNodeIP()
		assert.NoError(err)
		assert.NotEmpty(ip)
	}

	ip, err = stwrapper.getNextNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 255, 253}), ip)

	ip, err = stwrapper.getNextNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 255, 254}), ip)

	// 10.118.255.255 (broadcast IP) should not be returned
	ip, err = stwrapper.getNextNodeIP()
	assert.Error(err)
	assert.Empty(ip)

	// error should still persist
	ip, err = stwrapper.getNextNodeIP()
	assert.Error(err)
	assert.Empty(ip)
}

func TestPopNextFreeNodeIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutNextNodeIP(netip.AddrFrom4([4]byte{10, 118, 0, 11})))

	ip, err := stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 11}), ip)

	ip, err = stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 12}), ip)

	ip, err = stwrapper.PopNextFreeNodeIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 13}), ip)

	require.NoError(stwrapper.PutFreedNodeVPNIP("10.118.0.13"))
	require.NoError(stwrapper.PutFreedNodeVPNIP("10.118.0.12"))
	ipsInStore := map[netip.Addr]struct{}{
		netip.AddrFrom4([4]byte{10, 118, 0, 12}): {},
		netip.AddrFrom4([4]byte{10, 118, 0, 13}): {},
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
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 14}), ip)
}

func TestGenerateNextCoordinatorIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutNextCoordinatorIP(netip.AddrFrom4([4]byte{10, 118, 0, 1})))

	ip, err := stwrapper.getNextCoordinatorIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 1}), ip)

	ip, err = stwrapper.getNextCoordinatorIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 2}), ip)

	ip, err = stwrapper.getNextCoordinatorIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 3}), ip)

	for i := 0; i < 7; i++ {
		ip, err = stwrapper.getNextCoordinatorIP()
		assert.NoError(err)
		assert.NotEmpty(ip)
	}

	// 10.118.0.11 (first Node IP) should not be returned
	ip, err = stwrapper.getNextCoordinatorIP()
	assert.Error(err)
	assert.Empty(ip)

	// error should still persist
	ip, err = stwrapper.getNextCoordinatorIP()
	assert.Error(err)
	assert.Empty(ip)
}

func TestPopNextFreeCoordinatorIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutNextCoordinatorIP(netip.AddrFrom4([4]byte{10, 118, 0, 1})))

	ip, err := stwrapper.PopNextFreeCoordinatorIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 1}), ip)

	ip, err = stwrapper.PopNextFreeCoordinatorIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 2}), ip)

	ip, err = stwrapper.PopNextFreeCoordinatorIP()
	assert.NoError(err)
	assert.Equal(netip.AddrFrom4([4]byte{10, 118, 0, 3}), ip)

	for i := 0; i < 7; i++ {
		_, err = stwrapper.PopNextFreeCoordinatorIP()
		require.NoError(err)
	}

	ip, err = stwrapper.PopNextFreeCoordinatorIP()
	assert.Error(err)
	assert.Empty(ip)

	require.NoError(stwrapper.PutFreedCoordinatorVPNIP("10.118.0.3"))
	require.NoError(stwrapper.PutFreedCoordinatorVPNIP("10.118.0.2"))
	ipsInStore := map[netip.Addr]struct{}{
		netip.AddrFrom4([4]byte{10, 118, 0, 3}): {},
		netip.AddrFrom4([4]byte{10, 118, 0, 2}): {},
	}

	ip, err = stwrapper.PopNextFreeCoordinatorIP()
	assert.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.PopNextFreeCoordinatorIP()
	assert.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.PopNextFreeCoordinatorIP()
	assert.Error(err)
	assert.Equal(netip.Addr{}, ip)
}

func TestGetFreedVPNIP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stor := store.NewStdStore()
	stwrapper := StoreWrapper{Store: stor}
	require.NoError(stwrapper.PutFreedCoordinatorVPNIP("203.0.113.1"))
	require.NoError(stwrapper.PutFreedCoordinatorVPNIP("203.0.113.2"))
	ipsInStore := map[netip.Addr]struct{}{
		netip.AddrFrom4([4]byte{203, 0, 113, 1}): {},
		netip.AddrFrom4([4]byte{203, 0, 113, 2}): {},
	}

	ip, err := stwrapper.getFreedVPNIP(prefixFreeCoordinatorIPs)
	require.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.getFreedVPNIP(prefixFreeCoordinatorIPs)
	require.NoError(err)
	assert.Contains(ipsInStore, ip)
	delete(ipsInStore, ip)

	ip, err = stwrapper.getFreedVPNIP(prefixFreeCoordinatorIPs)
	var noElementsError *store.NoElementsLeftError
	assert.ErrorAs(err, &noElementsError)
	assert.Equal(netip.Addr{}, ip)
}
