package storewrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// variables which will be used as a store-prefix start with prefix[...].
// variables which will be used as a store-key start with key[...].
const (
	keyHighestAvailableCoordinatorIP = "highestAvailableCoordinatorIP"
	keyHighestAvailableNodeIP        = "highestAvailableNodeIP"
	keyKubernetesJoinCommand         = "kubeJoin"
	keyPeersResourceVersion          = "peersResourceVersion"
	keyMasterSecret                  = "masterSecret"
	keyKubeConfig                    = "kubeConfig"
	keyClusterID                     = "clusterID"
	keyKMSData                       = "KMSData"
	keyKEKID                         = "kekID"
	prefixFreeCoordinatorIPs         = "freeCoordinatorVPNIPs"
	prefixPeerLocation               = "peerPrefix"
	prefixFreeNodeIPs                = "freeNodeVPNIPs"
)

var (
	coordinatorIPRangeStart = netip.AddrFrom4([4]byte{10, 118, 0, 1})
	coordinatorIPRangeEnd   = netip.AddrFrom4([4]byte{10, 118, 0, 10})
	nodeIPRangeStart        = netip.AddrFrom4([4]byte{10, 118, 0, 11})
	nodeIPRangeEnd          = netip.AddrFrom4([4]byte{10, 118, 255, 254})
)

// StoreWrapper is a wrapper for the store interface.
type StoreWrapper struct {
	Store interface {
		Get(string) ([]byte, error)
		Put(string, []byte) error
		Delete(string) error
		Iterator(string) (store.Iterator, error)
	}
}

// GetState returns the state from store.
func (s StoreWrapper) GetState() (state.State, error) {
	rawState, err := s.Store.Get("state")
	if err != nil {
		return 0, err
	}

	currState, err := strconv.Atoi(string(rawState))
	if err != nil {
		return 0, err
	}

	return state.State(currState), nil
}

// PutState saves the state to store.
func (s StoreWrapper) PutState(currState state.State) error {
	rawState := []byte(strconv.Itoa(int(currState)))
	return s.Store.Put("state", rawState)
}

// PutPeer puts a single peer in the store, with a unique key derived form the VPNIP.
func (s StoreWrapper) PutPeer(peer peer.Peer) error {
	if len(peer.VPNIP) == 0 {
		return errors.New("unique ID of peer not set")
	}
	jsonPeer, err := json.Marshal(peer)
	if err != nil {
		return err
	}
	return s.Store.Put(prefixPeerLocation+peer.VPNIP, jsonPeer)
}

// RemovePeer removes a peer from the store.
func (s StoreWrapper) RemovePeer(peer peer.Peer) error {
	return s.Store.Delete(prefixPeerLocation + peer.VPNIP)
}

// GetPeers returns all peers in the store.
func (s StoreWrapper) GetPeers() ([]peer.Peer, error) {
	return s.getPeersByPrefix(prefixPeerLocation)
}

// IncrementPeersResourceVersion increments the version of the stored peers.
// Should be called in a transaction together with Add/Remove operation(s).
func (s StoreWrapper) IncrementPeersResourceVersion() error {
	val, err := s.GetPeersResourceVersion()
	var unsetErr *store.ValueUnsetError
	if errors.As(err, &unsetErr) {
		val = 0
	} else if err != nil {
		return err
	}
	return s.Store.Put(keyPeersResourceVersion, []byte(strconv.Itoa(val+1)))
}

// GetPeersResourceVersion returns the current version of the stored peers.
func (s StoreWrapper) GetPeersResourceVersion() (int, error) {
	raw, err := s.Store.Get(keyPeersResourceVersion)
	if err != nil {
		return 0, err
	}
	val, err := strconv.Atoi(string(raw))
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (s StoreWrapper) getPeersByPrefix(prefix string) ([]peer.Peer, error) {
	peerKeys, err := s.Store.Iterator(prefix)
	if err != nil {
		return nil, err
	}
	var peers []peer.Peer
	for peerKeys.HasNext() {
		storeKey, err := peerKeys.GetNext()
		if err != nil {
			return nil, err
		}
		marshalPeer, err := s.Store.Get(storeKey)
		if err != nil {
			return nil, err
		}
		var peer peer.Peer
		if err := json.Unmarshal(marshalPeer, &peer); err != nil {
			return nil, err
		}
		peers = append(peers, peer)

	}
	return peers, nil
}

// GetKubernetesJoinArgs returns the Kubernetes join command from store.
func (s StoreWrapper) GetKubernetesJoinArgs() (*kubeadm.BootstrapTokenDiscovery, error) {
	rawJoinCommand, err := s.Store.Get(keyKubernetesJoinCommand)
	if err != nil {
		return nil, err
	}
	joinCommand := kubeadm.BootstrapTokenDiscovery{}
	if err := json.Unmarshal(rawJoinCommand, &joinCommand); err != nil {
		return nil, err
	}
	return &joinCommand, nil
}

// PutKubernetesJoinArgs saves the Kubernetes join command to store.
func (s StoreWrapper) PutKubernetesJoinArgs(args *kubeadm.BootstrapTokenDiscovery) error {
	j, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return s.Store.Put(keyKubernetesJoinCommand, j)
}

// GetKubernetesConfig returns the Kubernetes kubeconfig file to authenticate with the Kubernetes API.
func (s StoreWrapper) GetKubernetesConfig() ([]byte, error) {
	return s.Store.Get(keyKubeConfig)
}

// PutKubernetesConfig saves the Kubernetes kubeconfig file command to store.
func (s StoreWrapper) PutKubernetesConfig(kubeConfig []byte) error {
	return s.Store.Put(keyKubeConfig, kubeConfig)
}

// GetMasterSecret returns the Constellation master secret from store.
func (s StoreWrapper) GetMasterSecret() ([]byte, error) {
	return s.Store.Get(keyMasterSecret)
}

// PutMasterSecret saves the Constellation master secret to store.
func (s StoreWrapper) PutMasterSecret(masterSecret []byte) error {
	return s.Store.Put(keyMasterSecret, masterSecret)
}

// GetKEKID returns the key encryption key ID from store.
func (s StoreWrapper) GetKEKID() (string, error) {
	kekID, err := s.Store.Get(keyKEKID)
	return string(kekID), err
}

// PutKEKID saves the key encryption key ID to store.
func (s StoreWrapper) PutKEKID(kekID string) error {
	return s.Store.Put(keyKEKID, []byte(kekID))
}

// GetKMSData returns the KMSData from the store.
func (s StoreWrapper) GetKMSData() (kms.KMSInformation, error) {
	storeData, err := s.Store.Get(keyKMSData)
	if err != nil {
		return kms.KMSInformation{}, err
	}
	data := kms.KMSInformation{}
	if err := json.Unmarshal(storeData, &data); err != nil {
		return kms.KMSInformation{}, err
	}
	return data, nil
}

// PutKMSData puts the KMSData in the store.
func (s StoreWrapper) PutKMSData(kmsInfo kms.KMSInformation) error {
	byteKMSInfo, err := json.Marshal(kmsInfo)
	if err != nil {
		return err
	}
	return s.Store.Put(keyKMSData, byteKMSInfo)
}

// GetClusterID returns the unique identifier of the cluster from store.
func (s StoreWrapper) GetClusterID() ([]byte, error) {
	return s.Store.Get(keyClusterID)
}

// PutClusterID saves the unique identifier of the cluster to store.
func (s StoreWrapper) PutClusterID(clusterID []byte) error {
	return s.Store.Put(keyClusterID, clusterID)
}

func (s StoreWrapper) InitializeStoreIPs() error {
	if err := s.PutNextNodeIP(nodeIPRangeStart); err != nil {
		return err
	}
	return s.PutNextCoordinatorIP(coordinatorIPRangeStart)
}

// PutNextCoordinatorIP puts the last used ip into the store.
func (s StoreWrapper) PutNextCoordinatorIP(ip netip.Addr) error {
	return s.Store.Put(keyHighestAvailableCoordinatorIP, ip.AsSlice())
}

// getNextCoordinatorIP generates addresses from a /16 subnet.
func (s StoreWrapper) getNextCoordinatorIP() (netip.Addr, error) {
	byteIP, err := s.Store.Get(keyHighestAvailableCoordinatorIP)
	if err != nil {
		return netip.Addr{}, errors.New("could not obtain IP from store")
	}
	ip, ok := netip.AddrFromSlice(byteIP)
	if !ok {
		return netip.Addr{}, fmt.Errorf("ip addr malformed %v", byteIP)
	}
	if !ip.IsValid() || ip.Compare(coordinatorIPRangeEnd) == 1 {
		return netip.Addr{}, errors.New("no ips left to assign")
	}
	nextIP := ip.Next()
	if err := s.PutNextCoordinatorIP(nextIP); err != nil {
		return netip.Addr{}, errors.New("could not put IP to store")
	}
	return ip, nil
}

// PopNextFreeCoordinatorIP return the next free IP, these could be a old one from a removed peer
// or a newly generated IP.
func (s StoreWrapper) PopNextFreeCoordinatorIP() (netip.Addr, error) {
	vpnIP, err := s.getFreedVPNIP(prefixFreeCoordinatorIPs)
	var noElementsError *store.NoElementsLeftError
	if errors.As(err, &noElementsError) {
		return s.getNextCoordinatorIP()
	}
	if err != nil {
		return netip.Addr{}, err
	}
	return vpnIP, nil
}

// PutFreedCoordinatorVPNIP puts a already generated VPNIP (IP < highestAvailableCoordinatorIP ),
// which is currently unused, into the store.
// The IP is saved at a specific prefix and will be used with priority when we
// request a new Coordinator IP.
func (s StoreWrapper) PutFreedCoordinatorVPNIP(vpnIP string) error {
	return s.Store.Put(prefixFreeCoordinatorIPs+vpnIP, nil)
}

// PutNextNodeIP puts the last used ip into the store.
func (s StoreWrapper) PutNextNodeIP(ip netip.Addr) error {
	return s.Store.Put(keyHighestAvailableNodeIP, ip.AsSlice())
}

// getNextNodeIP generates addresses from a /16 subnet.
func (s StoreWrapper) getNextNodeIP() (netip.Addr, error) {
	byteIP, err := s.Store.Get(keyHighestAvailableNodeIP)
	if err != nil {
		return netip.Addr{}, errors.New("could not obtain IP from store")
	}
	ip, ok := netip.AddrFromSlice(byteIP)
	if !ok {
		return netip.Addr{}, fmt.Errorf("ip addr malformed %v", byteIP)
	}
	if !ip.IsValid() || ip.Compare(nodeIPRangeEnd) == 1 {
		return netip.Addr{}, errors.New("no ips left to assign")
	}
	nextIP := ip.Next()

	if err := s.PutNextNodeIP(nextIP); err != nil {
		return netip.Addr{}, errors.New("could not put IP to store")
	}
	return ip, nil
}

// PopNextFreeNodeIP return the next free IP, these could be a old one from a removed peer
// or a newly generated IP.
func (s StoreWrapper) PopNextFreeNodeIP() (netip.Addr, error) {
	vpnIP, err := s.getFreedVPNIP(prefixFreeNodeIPs)
	var noElementsError *store.NoElementsLeftError
	if errors.As(err, &noElementsError) {
		return s.getNextNodeIP()
	}
	if err != nil {
		return netip.Addr{}, err
	}
	return vpnIP, nil
}

// PutFreedNodeVPNIP puts a already generated VPNIP (IP < highestAvailableNodeIP ),
// which is currently unused, into the store.
// The IP is saved at a specific prefix and will be used with priority when we
// request a new Node IP.
func (s StoreWrapper) PutFreedNodeVPNIP(vpnIP string) error {
	return s.Store.Put(prefixFreeNodeIPs+vpnIP, nil)
}

// getFreedVPNIP reclaims a VPNIP from the store and removes it from there.
func (s StoreWrapper) getFreedVPNIP(prefix string) (netip.Addr, error) {
	iter, err := s.Store.Iterator(prefix)
	if err != nil {
		return netip.Addr{}, err
	}
	vpnIPWithPrefix, err := iter.GetNext()
	if err != nil {
		return netip.Addr{}, err
	}
	stringVPNIP := strings.TrimPrefix(vpnIPWithPrefix, prefix)
	vpnIP, err := netip.ParseAddr(stringVPNIP)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("ip addr malformed %v, %w", stringVPNIP, err)
	}

	return vpnIP, s.Store.Delete(vpnIPWithPrefix)
}
