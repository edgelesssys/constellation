package pubapi

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/state"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

type Core interface {
	GetVPNPubKey() ([]byte, error)
	SetVPNIP(string) error
	GetCoordinatorVPNIP() string
	AddAdmin(pubKey []byte) (string, error)
	GenerateNextIP() (string, error)
	SwitchToPersistentStore() error
	GetIDs(masterSecret []byte) (ownerID []byte, clusterID []byte, err error)
	SetUpKMS(ctx context.Context, storageURI, kmsURI, kekID string, useExisting bool) error

	GetState() state.State
	RequireState(...state.State) error
	AdvanceState(newState state.State, ownerID, clusterID []byte) error

	GetPeers(resourceVersion int) (int, []peer.Peer, error)
	AddPeer(peer.Peer) error
	UpdatePeers([]peer.Peer) error

	InitCluster(autoscalingNodeGroups []string, cloudServiceAccountURI string) ([]byte, error)
	JoinCluster(kubeadm.BootstrapTokenDiscovery) error
}
