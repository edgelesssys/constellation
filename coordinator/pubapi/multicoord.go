package pubapi

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ActivateAsAdditionalCoordinator is the RPC call to activate subsequent coordinators.
func (a *API) ActivateAsAdditionalCoordinator(ctx context.Context, in *pubproto.ActivateAsAdditionalCoordinatorRequest) (out *pubproto.ActivateAsAdditionalCoordinatorResponse, reterr error) {
	_, cancel := context.WithTimeout(ctx, deadlineDuration)
	defer cancel()
	a.mut.Lock()
	defer a.mut.Unlock()

	if err := a.core.RequireState(state.AcceptingInit); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	// Some of the following actions can't be reverted (yet). If there's an
	// error, we may be in a weird state. Thus, mark this peer as failed.
	defer func() {
		if reterr != nil {
			_ = a.core.AdvanceState(state.Failed, nil, nil)
		}
	}()

	// AdvanceState MUST be called before any other functions that are not sanity checks or otherwise required
	// This ensures the node is marked as initialzed before the node is in a state that allows code execution
	// Any new additions to ActivateAsAdditionalCoordinator MUST come after
	if err := a.core.AdvanceState(state.ActivatingNodes, in.OwnerId, in.ClusterId); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// TODO: add KMS functions

	// add one coordinator to the VPN
	if err := a.core.SetVPNIP(in.AssignedVpnIp); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if err := a.core.AddPeerToVPN(peer.FromPubProto([]*pubproto.Peer{in.ActivatingCoordinatorData})[0]); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// run the VPN-API server
	if err := a.vpnAPIServer.Listen(net.JoinHostPort(in.AssignedVpnIp, vpnAPIPort)); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	a.wgClose.Add(1)
	go func() {
		defer a.wgClose.Done()
		if err := a.vpnAPIServer.Serve(); err != nil {
			panic(err)
		}
	}()

	// TODO: kubernetes information and join

	// ATTENTION: STORE HAS TO BE EMPTY (NO OVERLAPPING KEYS) WHEN THIS FUNCTION IS CALLED
	if err := a.core.SwitchToPersistentStore(); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	a.logger.Info("Transition to persistent store successful")

	// regularly get (peer) updates from etcd
	// start update before manual peer add to omit race conditions when multiple coordinator are activating nodes

	thisPeer, err := a.assemblePeerStruct(in.AssignedVpnIp, role.Coordinator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if err := a.core.AddPeerToStore(thisPeer); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	resourceVersion, peers, err := a.core.GetPeers(0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	a.resourceVersion = resourceVersion

	err = a.core.UpdatePeers(peers)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	// Manually trigger an update operation on all peers.
	// This may be expendable in the future, depending on whether it's acceptable that it takes
	// some seconds until the nodes get all peer data via their regular update requests.
	_, peers, err = a.core.GetPeers(0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	a.logger.Info("", zap.Any("peers", peers))
	for _, p := range peers {
		if p.Role == role.Node {
			if err := a.triggerNodeUpdate(p.PublicIP); err != nil {
				a.logger.Error("triggerNodeUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
		if p.Role == role.Coordinator && p.VPNIP != thisPeer.VPNIP {
			a.logger.Info("update coordinator", zap.String("coordinator vpnIP", p.VPNIP))
			if err := a.triggerCoordinatorUpdate(context.TODO(), p.PublicIP); err != nil {
				a.logger.Error("triggerCoordinatorUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
	}

	return &pubproto.ActivateAsAdditionalCoordinatorResponse{}, nil
}

func (a *API) ActivateAdditionalCoordinator(ctx context.Context, in *pubproto.ActivateAdditionalCoordinatorRequest) (*pubproto.ActivateAdditionalCoordinatorResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, deadlineDuration)
	defer cancel()

	if err := a.core.RequireState(state.ActivatingNodes); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	assignedVPNIP, err := a.core.GetNextCoordinatorIP()
	if err != nil {
		return nil, err
	}
	vpnIP, err := a.core.GetVPNIP()
	if err != nil {
		return nil, err
	}
	thisPeer, err := a.assemblePeerStruct(vpnIP, role.Coordinator)
	if err != nil {
		return nil, err
	}
	ownerID, clusterID, err := a.core.GetIDs(nil)
	if err != nil {
		return nil, err
	}

	conn, err := a.dial(ctx, net.JoinHostPort(in.CoordinatorPublicIp, endpointAVPNPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)

	_, err = client.ActivateAsAdditionalCoordinator(ctx, &pubproto.ActivateAsAdditionalCoordinatorRequest{
		AssignedVpnIp:             assignedVPNIP,
		ActivatingCoordinatorData: peer.ToPubProto([]peer.Peer{thisPeer})[0],
		OwnerId:                   ownerID,
		ClusterId:                 clusterID,
	})
	if err != nil {
		a.logger.Error("coordinator activation failed", zap.Error(err))
		return nil, err
	}

	return &pubproto.ActivateAdditionalCoordinatorResponse{}, nil
}

func (a *API) TriggerCoordinatorUpdate(ctx context.Context, in *pubproto.TriggerCoordinatorUpdateRequest) (*pubproto.TriggerCoordinatorUpdateResponse, error) {
	if err := a.core.RequireState(state.ActivatingNodes); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	resourceVersion, peers, err := a.core.GetPeers(a.resourceVersion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if resourceVersion == a.resourceVersion {
		a.logger.Info("coordinator: ressource version identical, no need to update")
		return &pubproto.TriggerCoordinatorUpdateResponse{}, nil
	}
	err = a.core.UpdatePeers(peers)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	a.resourceVersion = resourceVersion
	return &pubproto.TriggerCoordinatorUpdateResponse{}, nil
}

func (a *API) triggerCoordinatorUpdate(ctx context.Context, publicIP string) error {
	ctx, cancel := context.WithTimeout(ctx, deadlineDuration)
	defer cancel()

	// We don't verify the peer certificate here, since TriggerNodeUpdate triggers a connection over VPN
	// The target of the rpc needs to already be part of the VPN to process the request, meaning it is trusted
	conn, err := a.dialNoVerify(ctx, net.JoinHostPort(publicIP, endpointAVPNPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.TriggerCoordinatorUpdate(ctx, &pubproto.TriggerCoordinatorUpdateRequest{})

	return err
}
