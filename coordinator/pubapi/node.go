package pubapi

import (
	"context"
	"net"
	"time"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// ActivateAsNode is the RPC call to activate a Node.
func (a *API) ActivateAsNode(ctx context.Context, in *pubproto.ActivateAsNodeRequest) (resp *pubproto.ActivateAsNodeResponse, reterr error) {
	a.mut.Lock()
	defer a.mut.Unlock()

	if err := a.core.RequireState(state.AcceptingInit); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}

	if len(in.OwnerId) == 0 || len(in.ClusterId) == 0 {
		a.logger.Error("missing data to taint worker node as initialized")
		return nil, status.Error(codes.InvalidArgument, "missing data to taint worker node as initialized")
	}

	// If any of the following actions fail, we cannot revert.
	// Thus, mark this peer as failed.
	defer func() {
		if reterr != nil {
			_ = a.core.AdvanceState(state.Failed, nil, nil)
		}
	}()

	// AdvanceState MUST be called before any other functions that are not sanity checks or otherwise required
	// This ensures the node is marked as initialzed before the node is in a state that allows code execution
	// Any new additions to ActivateAsNode MUST come after
	if err := a.core.AdvanceState(state.NodeWaitingForClusterJoin, in.OwnerId, in.ClusterId); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	vpnPubKey, err := a.core.GetVPNPubKey()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if err := a.core.SetVPNIP(in.NodeVpnIp); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// add initial peers
	if err := a.core.UpdatePeers(peer.FromPubProto(in.Peers)); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// persist node state on disk
	if err := a.core.PersistNodeState(role.Node, in.OwnerId, in.ClusterId); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// regularly get (peer) updates from Coordinator
	a.wgClose.Add(1)
	go a.updateLoop()

	return &pubproto.ActivateAsNodeResponse{NodeVpnPubKey: vpnPubKey}, nil
}

// JoinCluster is the RPC call to request this node to join the cluster.
func (a *API) JoinCluster(ctx context.Context, in *pubproto.JoinClusterRequest) (*pubproto.JoinClusterResponse, error) {
	a.mut.Lock()
	defer a.mut.Unlock()

	if err := a.core.RequireState(state.NodeWaitingForClusterJoin); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}

	conn, err := a.dialInsecure(ctx, net.JoinHostPort(in.CoordinatorVpnIp, vpnAPIPort))
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "%v", err)
	}
	resp, err := vpnproto.NewAPIClient(conn).GetK8SJoinArgs(ctx, &vpnproto.GetK8SJoinArgsRequest{})
	conn.Close()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	err = a.core.JoinCluster(kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: resp.ApiServerEndpoint,
		Token:             resp.Token,
		CACertHashes:      []string{resp.DiscoveryTokenCaCertHash},
	})
	if err != nil {
		_ = a.core.AdvanceState(state.Failed, nil, nil)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if err := a.core.AdvanceState(state.IsNode, nil, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pubproto.JoinClusterResponse{}, nil
}

// TriggerNodeUpdate is the RPC call to request this node to get an update from the Coordinator.
func (a *API) TriggerNodeUpdate(ctx context.Context, in *pubproto.TriggerNodeUpdateRequest) (*pubproto.TriggerNodeUpdateResponse, error) {
	if err := a.core.RequireState(state.IsNode); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	if err := a.update(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pubproto.TriggerNodeUpdateResponse{}, nil
}

func (a *API) updateLoop() {
	defer a.wgClose.Done()
	ticker := time.NewTicker(updateInterval)

	for {
		if err := a.update(context.Background()); err != nil {
			a.logger.Error("updateLoop: update failed", zap.Error(err))
		}
		select {
		case <-a.stopUpdate:
			ticker.Stop()
			return
		case <-ticker.C:
		}
	}
}

func (a *API) update(ctx context.Context) error {
	a.mut.Lock()
	defer a.mut.Unlock()

	ctx, cancel := context.WithTimeout(ctx, deadlineDuration)
	defer cancel()

	// TODO: replace hardcoded IP
	conn, err := a.dialInsecure(ctx, net.JoinHostPort("10.118.0.1", vpnAPIPort))
	if err != nil {
		return err
	}
	resp, err := vpnproto.NewAPIClient(conn).GetUpdate(ctx, &vpnproto.GetUpdateRequest{ResourceVersion: int64(a.resourceVersion)})
	conn.Close()
	if err != nil {
		return err
	}

	resourceVersion := int(resp.ResourceVersion)
	if resourceVersion == a.resourceVersion {
		return nil
	}

	// TODO does this naive approach of performing full updates everytime need to be replaced by something more clever like watches in K8s?
	// https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes

	if err := a.core.UpdatePeers(peer.FromVPNProto(resp.Peers)); err != nil {
		return err
	}
	a.resourceVersion = resourceVersion

	return nil
}
