package pubapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/state"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

type fakeCore struct {
	vpnPubKey                  []byte
	vpnIP                      string
	setVPNIPErr                error
	adminPubKey                []byte
	nextIP                     int
	switchToPersistentStoreErr error
	state                      state.State
	ownerID                    []byte
	clusterID                  []byte
	peers                      []peer.Peer
	updatedPeers               [][]peer.Peer
	kubeconfig                 []byte
	autoscalingNodeGroups      []string
	joinArgs                   []kubeadm.BootstrapTokenDiscovery
	joinClusterErr             error
	kekID                      string
}

func (c *fakeCore) GetVPNPubKey() ([]byte, error) {
	return c.vpnPubKey, nil
}

func (c *fakeCore) SetVPNIP(ip string) error {
	if len(c.ownerID) == 0 || len(c.clusterID) == 0 {
		return errors.New("SetVPNIP called before IDs were set")
	}
	c.vpnIP = ip
	return c.setVPNIPErr
}

func (*fakeCore) GetCoordinatorVPNIP() string {
	return "192.0.2.100"
}

func (c *fakeCore) AddAdmin(pubKey []byte) (string, error) {
	c.adminPubKey = pubKey
	return "192.0.2.99", nil
}

func (c *fakeCore) GenerateNextIP() (string, error) {
	c.nextIP++
	return fmt.Sprintf("192.0.2.%v", 100+c.nextIP), nil
}

func (c *fakeCore) SwitchToPersistentStore() error {
	return c.switchToPersistentStoreErr
}

func (c *fakeCore) GetIDs(masterSecret []byte) (ownerID []byte, clusterID []byte, err error) {
	return c.ownerID, c.clusterID, nil
}

func (c *fakeCore) GetState() state.State {
	return c.state.Get()
}

func (c *fakeCore) RequireState(states ...state.State) error {
	return c.state.Require(states...)
}

func (c *fakeCore) AdvanceState(newState state.State, ownerID, clusterID []byte) error {
	c.ownerID = ownerID
	c.clusterID = clusterID
	c.state.Advance(newState)
	return nil
}

func (*fakeCore) GetPeers(resourceVersion int) (int, []peer.Peer, error) {
	return 0, nil, nil
}

func (c *fakeCore) AddPeer(peer peer.Peer) error {
	c.peers = append(c.peers, peer)
	return nil
}

func (c *fakeCore) UpdatePeers(peers []peer.Peer) error {
	c.updatedPeers = append(c.updatedPeers, peers)
	return nil
}

func (c *fakeCore) InitCluster(autoscalingNodeGroups []string) ([]byte, error) {
	c.autoscalingNodeGroups = autoscalingNodeGroups
	return c.kubeconfig, nil
}

func (c *fakeCore) JoinCluster(args kubeadm.BootstrapTokenDiscovery) error {
	c.joinArgs = append(c.joinArgs, args)
	return c.joinClusterErr
}

func (c *fakeCore) SetUpKMS(ctx context.Context, storageURI, kmsURI, kekID string, useExisting bool) error {
	c.kekID = kekID
	return nil
}
