package storewrapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/store"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

const (
	requestMasterSecret          = "masterSecret"
	requestClusterID             = "clusterID"
	requestVPNPubKey             = "vpnKey"
	requestKubernetesJoinCommand = "kubeJoin"
	requestKubeConfig            = "kubeConfig"
	requestKEKID                 = "kekID"
	peerLocationPrefix           = "PeerPrefix"
	peersResourceVersion         = "peersResourceVersion"
	adminLocation                = "externalAdmin"
	freeNodeIP                   = "freeNodeVPNIPs"
	lastNodeIP                   = "LastNodeIPPrefix"
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

// GetVPNKey returns the VPN pubKey from Store.
func (s StoreWrapper) GetVPNKey() ([]byte, error) {
	return s.Store.Get(requestVPNPubKey)
}

// PutVPNKey saves the VPN pubKey to store.
func (s StoreWrapper) PutVPNKey(key []byte) error {
	return s.Store.Put(requestVPNPubKey, key)
}

// PutPeer puts a single peer in the store, with a unique key derived form the VPNIP.
func (s StoreWrapper) PutPeer(peer peer.Peer) error {
	if len(peer.VPNIP) == 0 {
		return fmt.Errorf("unique ID of peer not set")
	}
	jsonPeer, err := json.Marshal(peer)
	if err != nil {
		return err
	}
	return s.Store.Put(peerLocationPrefix+peer.VPNIP, jsonPeer)
}

// RemovePeer removes a peer from the store.
func (s StoreWrapper) RemovePeer(peer peer.Peer) error {
	return s.Store.Delete(peerLocationPrefix + peer.VPNIP)
}

// GetPeer returns a peer requested by the given VPN IP address.
func (s StoreWrapper) GetPeer(vpnIP string) (peer.Peer, error) {
	bytePeer, err := s.Store.Get(peerLocationPrefix + vpnIP)
	if err != nil {
		return peer.Peer{}, err
	}
	var peer peer.Peer
	err = json.Unmarshal(bytePeer, &peer)
	return peer, err
}

// GetPeers returns all peers in the store.
func (s StoreWrapper) GetPeers() ([]peer.Peer, error) {
	return s.getPeersByPrefix(peerLocationPrefix)
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
	return s.Store.Put(peersResourceVersion, []byte(strconv.Itoa(val+1)))
}

// GetPeersResourceVersion returns the current version of the stored peers.
func (s StoreWrapper) GetPeersResourceVersion() (int, error) {
	raw, err := s.Store.Get(peersResourceVersion)
	if err != nil {
		return 0, err
	}
	val, err := strconv.Atoi(string(raw))
	if err != nil {
		return 0, err
	}
	return val, nil
}

// UpdatePeers synchronizes the stored peers with the passed peers, returning added and removed peers.
func (s StoreWrapper) UpdatePeers(peers []peer.Peer) (added, removed []peer.Peer, err error) {
	// convert to map for easier lookup
	updatedPeers := make(map[string]peer.Peer)
	for _, p := range peers {
		updatedPeers[p.VPNIP] = p
	}

	it, err := s.Store.Iterator(peerLocationPrefix)
	if err != nil {
		return nil, nil, err
	}

	// collect peers that need to be added or removed
	for it.HasNext() {
		key, err := it.GetNext()
		if err != nil {
			return nil, nil, err
		}
		val, err := s.Store.Get(key)
		if err != nil {
			return nil, nil, err
		}
		var storedPeer peer.Peer
		if err := json.Unmarshal(val, &storedPeer); err != nil {
			return nil, nil, err
		}

		if updPeer, ok := updatedPeers[storedPeer.VPNIP]; ok {
			if updPeer.PublicEndpoint != storedPeer.PublicEndpoint || !bytes.Equal(updPeer.VPNPubKey, storedPeer.VPNPubKey) {
				// stored peer must be updated, so mark for addition AND removal
				added = append(added, updPeer)
				removed = append(removed, storedPeer)
			}
			delete(updatedPeers, updPeer.VPNIP)
		} else {
			// stored peer is not contained in the updated peers, so mark for removal
			removed = append(removed, storedPeer)
		}
	}

	// remaining updated peers were not in the store, so mark for addition
	for _, p := range updatedPeers {
		added = append(added, p)
	}

	// perform remove and add
	for _, p := range removed {
		if err := s.Store.Delete(peerLocationPrefix + p.VPNIP); err != nil {
			return nil, nil, err
		}
	}
	for _, p := range added {
		data, err := json.Marshal(p)
		if err != nil {
			return nil, nil, err
		}
		if err := s.Store.Put(peerLocationPrefix+p.VPNIP, data); err != nil {
			return nil, nil, err
		}
	}

	return added, removed, nil
}

// PutAdmin puts a single admin in the store, with a unique key derived form the VPNIP.
func (s StoreWrapper) PutAdmin(peer peer.AdminData) error {
	jsonPeer, err := json.Marshal(peer)
	if err != nil {
		return err
	}
	return s.Store.Put(adminLocation+peer.VPNIP, jsonPeer)
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

// PutFreedNodeVPNIP stores a VPNIP in the store at a unique key.
func (s StoreWrapper) PutFreedNodeVPNIP(vpnIP string) error {
	return s.Store.Put(freeNodeIP+vpnIP, nil)
}

// GetFreedNodeVPNIP reclaims a VPNIP from the store and removes it from there.
func (s StoreWrapper) GetFreedNodeVPNIP() (string, error) {
	iter, err := s.Store.Iterator(freeNodeIP)
	if err != nil {
		return "", err
	}
	if !iter.HasNext() {
		return "", nil
	}
	vpnIP, err := iter.GetNext()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(vpnIP, freeNodeIP), s.Store.Delete(vpnIP)
}

// GetKubernetesJoinArgs returns the Kubernetes join command from store.
func (s StoreWrapper) GetKubernetesJoinArgs() (*kubeadm.BootstrapTokenDiscovery, error) {
	rawJoinCommand, err := s.Store.Get(requestKubernetesJoinCommand)
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
	return s.Store.Put(requestKubernetesJoinCommand, j)
}

// GetKubernetesConfig returns the Kubernetes kubeconfig file to authenticate with the kubernetes API.
func (s StoreWrapper) GetKubernetesConfig() ([]byte, error) {
	return s.Store.Get(requestKubeConfig)
}

// PutKubernetesConfig saves the Kubernetes kubeconfig file command to store.
func (s StoreWrapper) PutKubernetesConfig(kubeConfig []byte) error {
	return s.Store.Put(requestKubeConfig, kubeConfig)
}

// GetMasterSecret returns the Constellation master secret from store.
func (s StoreWrapper) GetMasterSecret() ([]byte, error) {
	return s.Store.Get(requestMasterSecret)
}

// PutMasterSecret saves the Constellation master secret to store.
func (s StoreWrapper) PutMasterSecret(masterSecret []byte) error {
	return s.Store.Put(requestMasterSecret, masterSecret)
}

// GetKEKID returns the key encryption key ID from store.
func (s StoreWrapper) GetKEKID() (string, error) {
	kekID, err := s.Store.Get(requestKEKID)
	return string(kekID), err
}

// PutKEKID saves the key encryption key ID to store.
func (s StoreWrapper) PutKEKID(kekID string) error {
	return s.Store.Put(requestKEKID, []byte(kekID))
}

// GetClusterID returns the unique identifier of the cluster from store.
func (s StoreWrapper) GetClusterID() ([]byte, error) {
	return s.Store.Get(requestClusterID)
}

// PutClusterID saves the unique identifier of the cluster to store.
func (s StoreWrapper) PutClusterID(clusterID []byte) error {
	return s.Store.Put(requestClusterID, clusterID)
}

// PutLastNodeIP puts the last used ip to the store.
func (s StoreWrapper) PutLastNodeIP(ip []byte) error {
	return s.Store.Put(lastNodeIP, ip)
}

// GetLastNodeIP gets the last used ip from the store.
func (s StoreWrapper) GetLastNodeIP() ([]byte, error) {
	return s.Store.Get(lastNodeIP)
}

// generateNextNodeIP generates addresses from a /16 subnet.
func (s StoreWrapper) generateNextNodeIP() (string, error) {
	ip, err := s.GetLastNodeIP()
	if err != nil {
		return "", fmt.Errorf("could not obtain IP from store")
	}
	if ip[3] < 255 && ip[2] < 255 || ip[3] < 254 {
		ip[3]++
	} else if ip[2] < 255 {
		ip[3] = 0
		ip[2]++
	} else {
		return "", fmt.Errorf("no IPs left to assign")
	}

	err = s.PutLastNodeIP(ip)
	if err != nil {
		return "", fmt.Errorf("could not put IP to store")
	}
	return net.IPv4(ip[0], ip[1], ip[2], ip[3]).String(), nil
}

// PopNextFreeNodeIP return the next free IP, these could be a old one from a removed peer
// or a newly generated IP.
func (s StoreWrapper) PopNextFreeNodeIP() (string, error) {
	vpnIP, err := s.GetFreedNodeVPNIP()
	if err != nil {
		return "", err
	}
	if len(vpnIP) == 0 {
		return s.generateNextNodeIP()
	}
	return vpnIP, nil
}
