// Package vpnapi implements the API that a coordinator exposes inside the VPN.
package vpnapi

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	gpeer "google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// API is the API.
type API struct {
	logger *zap.Logger
	core   Core
	vpnproto.UnimplementedAPIServer
}

// New creates a new API.
func New(logger *zap.Logger, core Core) *API {
	return &API{
		logger: logger,
		core:   core,
	}
}

// GetUpdate returns updated information to a node. It also recognizes the call as a node's heartbeat.
func (a *API) GetUpdate(ctx context.Context, in *vpnproto.GetUpdateRequest) (*vpnproto.GetUpdateResponse, error) {
	if client, ok := gpeer.FromContext(ctx); ok {
		a.core.NotifyNodeHeartbeat(client.Addr)
	} else {
		a.logger.DPanic("Failed to get peer info from context.")
	}

	resourceVersion, peers, err := a.core.GetPeers(int(in.ResourceVersion))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &vpnproto.GetUpdateResponse{ResourceVersion: int64(resourceVersion), Peers: peer.ToVPNProto(peers)}, nil
}

// GetK8SJoinArgs is the RPC call to get the K8s join args.
func (a *API) GetK8SJoinArgs(ctx context.Context, in *vpnproto.GetK8SJoinArgsRequest) (*vpnproto.GetK8SJoinArgsResponse, error) {
	args, err := a.core.GetK8sJoinArgs()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &vpnproto.GetK8SJoinArgsResponse{
		ApiServerEndpoint:        args.APIServerEndpoint,
		Token:                    args.Token,
		DiscoveryTokenCaCertHash: args.CACertHashes[0],
	}, nil
}

// GetDataKey returns a data key derived from the Constellation's master secret.
func (a *API) GetDataKey(ctx context.Context, in *vpnproto.GetDataKeyRequest) (*vpnproto.GetDataKeyResponse, error) {
	key, err := a.core.GetDataKey(ctx, in.DataKeyId, int(in.Length))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &vpnproto.GetDataKeyResponse{DataKey: key}, nil
}

type Core interface {
	GetPeers(resourceVersion int) (int, []peer.Peer, error)
	NotifyNodeHeartbeat(net.Addr)
	GetK8sJoinArgs() (*kubeadm.BootstrapTokenDiscovery, error)
	GetDataKey(ctx context.Context, dataKeyID string, length int) ([]byte, error)
}
