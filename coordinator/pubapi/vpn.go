package pubapi

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetVPNPeers retrieves VPN peers from a coordinator.
func (a *API) GetVPNPeers(context.Context, *pubproto.GetVPNPeersRequest) (*pubproto.GetVPNPeersResponse, error) {
	_, peers, err := a.core.GetPeers(0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get peers: %v", err)
	}

	return &pubproto.GetVPNPeersResponse{
		Peers: peer.ToPubProto(peers),
	}, nil
}
