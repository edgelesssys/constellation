package pubapi

import (
	"context"
	"errors"
	"net/netip"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	kms "github.com/edgelesssys/constellation/kms/server/setup"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

type fakeCore struct {
	vpnPubKey                  []byte
	getvpnPubKeyErr            error
	vpnIP                      string
	setVPNIPErr                error
	nextNodeIP                 netip.Addr
	nextCoordinatorIP          netip.Addr
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
	UpdatePeersErr             error
	GetPeersErr                error
	persistNodeStateRoles      []role.Role
	persistNodeStateErr        error
	kekID                      string
	dataKey                    []byte
	getDataKeyErr              error
	linuxUserManager           user.LinuxUserManager
}

func (c *fakeCore) GetVPNPubKey() ([]byte, error) {
	return c.vpnPubKey, c.getvpnPubKeyErr
}

func (c *fakeCore) SetVPNIP(ip string) error {
	if len(c.ownerID) == 0 || len(c.clusterID) == 0 {
		return errors.New("SetVPNIP called before IDs were set")
	}
	c.vpnIP = ip
	return c.setVPNIPErr
}

func (c *fakeCore) InitializeStoreIPs() error {
	c.nextCoordinatorIP = netip.AddrFrom4([4]byte{10, 118, 0, 1})
	c.nextNodeIP = netip.AddrFrom4([4]byte{10, 118, 0, 11})
	return nil
}

func (c *fakeCore) GetVPNIP() (string, error) {
	return c.vpnIP, nil
}

func (c *fakeCore) GetNextNodeIP() (string, error) {
	ip := c.nextNodeIP.String()
	c.nextNodeIP = c.nextNodeIP.Next()
	return ip, nil
}

func (c *fakeCore) GetNextCoordinatorIP() (string, error) {
	ip := c.nextCoordinatorIP.String()
	c.nextCoordinatorIP = c.nextCoordinatorIP.Next()
	return ip, nil
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

func (c *fakeCore) GetPeers(resourceVersion int) (int, []peer.Peer, error) {
	return 1, c.peers, c.GetPeersErr
}

func (c *fakeCore) AddPeer(peer peer.Peer) error {
	c.peers = append(c.peers, peer)
	return nil
}

func (c *fakeCore) AddPeerToStore(peer peer.Peer) error {
	c.peers = append(c.peers, peer)
	return nil
}

func (c *fakeCore) AddPeerToVPN(peer peer.Peer) error {
	c.peers = append(c.peers, peer)
	return nil
}

func (c *fakeCore) UpdatePeers(peers []peer.Peer) error {
	c.updatedPeers = append(c.updatedPeers, peers)
	return c.UpdatePeersErr
}

func (c *fakeCore) InitCluster(autoscalingNodeGroups []string, cloudServiceAccountURI string, masterSecret []byte) ([]byte, error) {
	c.autoscalingNodeGroups = autoscalingNodeGroups
	return c.kubeconfig, nil
}

func (c *fakeCore) JoinCluster(args *kubeadm.BootstrapTokenDiscovery, _ string, _ role.Role) error {
	c.joinArgs = append(c.joinArgs, *args)
	return c.joinClusterErr
}

func (c *fakeCore) PersistNodeState(role role.Role, vpnIP string, ownerID []byte, clusterID []byte) error {
	c.persistNodeStateRoles = append(c.persistNodeStateRoles, role)
	return c.persistNodeStateErr
}

func (c *fakeCore) SetUpKMS(ctx context.Context, storageURI, kmsURI, kekID string, useExisting bool) error {
	c.kekID = kekID
	return nil
}

func (c *fakeCore) GetKMSInfo() (kms.KMSInformation, error) {
	return kms.KMSInformation{}, nil
}

func (c *fakeCore) GetDataKey(ctx context.Context, keyID string, length int) ([]byte, error) {
	return c.dataKey, c.getDataKeyErr
}

func (c *fakeCore) GetDiskUUID() (string, error) {
	return "fake-disk-uuid", nil
}

func (c *fakeCore) UpdateDiskPassphrase(passphrase string) error {
	return nil
}

func (c *fakeCore) CreateSSHUsers(sshUserKeys []ssh.UserKey) error {
	sshAccess := ssh.NewSSHAccess(c.linuxUserManager)
	ctx := context.Background()

	for _, pair := range sshUserKeys {
		if err := sshAccess.DeploySSHAuthorizedKey(ctx, pair); err != nil {
			return err
		}
	}

	return nil
}
