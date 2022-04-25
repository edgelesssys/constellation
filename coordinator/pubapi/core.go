package pubapi

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

type Core interface {
	GetVPNPubKey() ([]byte, error)
	SetVPNIP(string) error
	GetVPNIP() (string, error)
	InitializeStoreIPs() error
	GetNextNodeIP() (string, error)
	GetNextCoordinatorIP() (string, error)
	SwitchToPersistentStore() error
	GetIDs(masterSecret []byte) (ownerID []byte, clusterID []byte, err error)
	PersistNodeState(role role.Role, ownerID []byte, clusterID []byte) error
	SetUpKMS(ctx context.Context, storageURI, kmsURI, kekID string, useExisting bool) error
	GetKMSInfo() (kms.KMSInformation, error)
	GetDataKey(ctx context.Context, keyID string, length int) ([]byte, error)
	GetDiskUUID() (string, error)
	UpdateDiskPassphrase(passphrase string) error

	GetState() state.State
	RequireState(...state.State) error
	AdvanceState(newState state.State, ownerID, clusterID []byte) error

	GetPeers(resourceVersion int) (int, []peer.Peer, error)
	AddPeer(peer.Peer) error
	AddPeerToStore(peer.Peer) error
	AddPeerToVPN(peer.Peer) error
	UpdatePeers([]peer.Peer) error

	InitCluster(autoscalingNodeGroups []string, cloudServiceAccountURI string) ([]byte, error)
	JoinCluster(joinToken *kubeadm.BootstrapTokenDiscovery, certificateKey string, role role.Role) error
}
