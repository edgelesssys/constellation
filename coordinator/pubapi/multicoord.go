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
		return nil, status.Errorf(codes.FailedPrecondition, "node is not in required state: %v", err)
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
		return nil, status.Errorf(codes.Internal, "advance state to ActivatingNodes: %v", err)
	}
	// add one coordinator to the VPN
	if err := a.core.SetVPNIP(in.AssignedVpnIp); err != nil {
		return nil, status.Errorf(codes.Internal, "set vpn IP address: %v", err)
	}

	if err := a.core.AddPeerToVPN(peer.FromPubProto([]*pubproto.Peer{in.ActivatingCoordinatorData})[0]); err != nil {
		return nil, status.Errorf(codes.Internal, "adding initial peers to vpn: %v", err)
	}

	// run the VPN-API server
	if err := a.vpnAPIServer.Listen(net.JoinHostPort(in.AssignedVpnIp, vpnAPIPort)); err != nil {
		return nil, status.Errorf(codes.Internal, "start vpnAPIServer: %v", err)
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
		return nil, status.Errorf(codes.Internal, "switch to persistent store: %v", err)
	}
	a.logger.Info("Transition to persistent store successful")

	kmsData, err := a.core.GetKMSInfo()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if err := a.core.SetUpKMS(ctx, kmsData.StorageUri, kmsData.KmsUri, kmsData.KeyEncryptionKeyID, false); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// persist node state on disk
	if err := a.core.PersistNodeState(role.Coordinator, in.OwnerId, in.ClusterId); err != nil {
		return nil, status.Errorf(codes.Internal, "persist node state: %v", err)
	}
	diskUUID, err := a.core.GetDiskUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting disk uuid: %v", err)
	}
	diskKey, err := a.core.GetDataKey(ctx, diskUUID, 32)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting disk key: %v", err)
	}
	if err := a.core.UpdateDiskPassphrase(string(diskKey)); err != nil {
		return nil, status.Errorf(codes.Internal, "updating disk key: %v", err)
	}

	// regularly get (peer) updates from etcd
	// start update before manual peer add to omit race conditions when multiple coordinator are activating nodes

	thisPeer, err := a.assemblePeerStruct(in.AssignedVpnIp, role.Coordinator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "assembling coordinator peer struct: %v", err)
	}
	if err := a.core.AddPeerToStore(thisPeer); err != nil {
		return nil, status.Errorf(codes.Internal, "adding new coordinator to persistent store: %v", err)
	}

	resourceVersion, peers, err := a.core.GetPeers(0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get peers from store: %v", err)
	}
	a.resourceVersion = resourceVersion

	err = a.core.UpdatePeers(peers)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "synchronizing peers with vpn state: %v", err)
	}
	// Manually trigger an update operation on all peers.
	// This may be expendable in the future, depending on whether it's acceptable that it takes
	// some seconds until the nodes get all peer data via their regular update requests.
	_, peers, err = a.core.GetPeers(0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get peers from store: %v", err)
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
		return nil, status.Errorf(codes.FailedPrecondition, "coordinator is not in required state: %v", err)
	}
	assignedVPNIP, err := a.core.GetNextCoordinatorIP()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "requesting new coordinator vpn IP address: %v", err)
	}
	vpnIP, err := a.core.GetVPNIP()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get own vpn IP address: %v", err)
	}
	thisPeer, err := a.assemblePeerStruct(vpnIP, role.Coordinator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "assembling coordinator peer struct: %v", err)
	}
	ownerID, clusterID, err := a.core.GetIDs(nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get owner and cluster ID: %v", err)
	}

	conn, err := a.dial(ctx, net.JoinHostPort(in.CoordinatorPublicIp, endpointAVPNPort))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "dialing new coordinator: %v", err)
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
		return nil, status.Errorf(codes.Internal, "activate new coordinator: %v", err)
	}

	return &pubproto.ActivateAdditionalCoordinatorResponse{}, nil
}

func (a *API) TriggerCoordinatorUpdate(ctx context.Context, in *pubproto.TriggerCoordinatorUpdateRequest) (*pubproto.TriggerCoordinatorUpdateResponse, error) {
	if err := a.core.RequireState(state.ActivatingNodes); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "coordinator is not in required state for updating state: %v", err)
	}
	resourceVersion, peers, err := a.core.GetPeers(a.resourceVersion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get peers from store: %v", err)
	}
	if resourceVersion == a.resourceVersion {
		a.logger.Info("ressource version identical, no need to update")
		return &pubproto.TriggerCoordinatorUpdateResponse{}, nil
	}
	err = a.core.UpdatePeers(peers)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "synchronizing peers with vpn state: %v", err)
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
