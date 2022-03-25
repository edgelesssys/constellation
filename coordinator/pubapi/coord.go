package pubapi

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ActivateAsCoordinator is the RPC call to activate the Coordinator.
func (a *API) ActivateAsCoordinator(in *pubproto.ActivateAsCoordinatorRequest, srv pubproto.API_ActivateAsCoordinatorServer) (reterr error) {
	a.mut.Lock()
	defer a.mut.Unlock()

	if err := a.core.RequireState(state.AcceptingInit); err != nil {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}

	if len(in.MasterSecret) == 0 {
		a.logger.Error("missing master secret")
		return status.Error(codes.InvalidArgument, "missing master secret")
	}

	// If any of the following actions fail, we cannot revert
	// Thus, mark this peer as failed.
	defer func() {
		if reterr != nil {
			_ = a.core.AdvanceState(state.Failed, nil, nil)
		}
	}()

	// AdvanceState MUST be called before any other functions that are not sanity checks or otherwise required
	// This ensures the node is marked as initialzed before the node is in a state that allows code execution
	// Any new additions to ActivateAsNode MUST come after
	ownerID, clusterID, err := a.core.GetIDs(in.MasterSecret)
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}
	if err := a.core.AdvanceState(state.ActivatingNodes, ownerID, clusterID); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := a.core.SetUpKMS(ctx, in.StorageUri, in.KmsUri, in.KeyEncryptionKeyId, in.UseExistingKek); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	coordPeer, err := a.makeCoordinatorPeer()
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	if err := a.core.SetVPNIP(coordPeer.VPNIP); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}
	if err := a.core.AddPeer(coordPeer); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	kubeconfig, err := a.core.InitCluster(in.AutoscalingNodeGroups, in.CloudServiceAccountUri)
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	// run the VPN-API server
	if err := a.vpnAPIServer.Listen(net.JoinHostPort(coordPeer.VPNIP, vpnAPIPort)); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}
	a.wgClose.Add(1)
	go func() {
		defer a.wgClose.Done()
		if err := a.vpnAPIServer.Serve(); err != nil {
			panic(err)
		}
	}()

	logToCLI := a.newLogToCLIFunc(func(msg string) error {
		return srv.Send(&pubproto.ActivateAsCoordinatorResponse{
			Content: &pubproto.ActivateAsCoordinatorResponse_Log{
				Log: &pubproto.Log{
					Message: msg,
				},
			},
		})
	})

	// TODO: check performance and maybe make concurrent
	if err := a.activateNodes(logToCLI, in.NodePublicEndpoints, coordPeer); err != nil {
		a.logger.Error("node activation failed", zap.Error(err))
		return status.Errorf(codes.Internal, "%v", err)
	}

	if err := a.core.SwitchToPersistentStore(); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	// This effectively gives code execution, so we do this last.
	adminVPNIP, err := a.core.AddAdmin(in.AdminVpnPubKey)
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	return srv.Send(&pubproto.ActivateAsCoordinatorResponse{
		Content: &pubproto.ActivateAsCoordinatorResponse_AdminConfig{
			AdminConfig: &pubproto.AdminConfig{
				AdminVpnIp:           adminVPNIP,
				CoordinatorVpnPubKey: coordPeer.VPNPubKey,
				Kubeconfig:           kubeconfig,
				OwnerId:              ownerID,
				ClusterId:            clusterID,
			},
		},
	})
}

// ActivateAdditionalNodes is the RPC call to activate additional nodes.
func (a *API) ActivateAdditionalNodes(in *pubproto.ActivateAdditionalNodesRequest, srv pubproto.API_ActivateAdditionalNodesServer) error {
	if err := a.core.RequireState(state.ActivatingNodes); err != nil {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}

	coordPeer, err := a.makeCoordinatorPeer()
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	logToCLI := a.newLogToCLIFunc(func(msg string) error {
		return srv.Send(&pubproto.ActivateAdditionalNodesResponse{
			Log: &pubproto.Log{
				Message: msg,
			},
		})
	})

	// TODO: check performance and maybe make concurrent
	if err := a.activateNodes(logToCLI, in.NodePublicEndpoints, coordPeer); err != nil {
		a.logger.Error("node activation failed", zap.Error(err))
		return status.Errorf(codes.Internal, "%v", err)
	}

	return srv.Send(&pubproto.ActivateAdditionalNodesResponse{
		Log: &pubproto.Log{
			Message: "success",
		},
	})
}

func (a *API) activateNodes(logToCLI logFunc, nodePublicEndpoints []string, coordPeer peer.Peer) error {
	// Create initial peer data to be sent to the nodes. Currently, this is just this Coordinator.
	initialPeers := peer.ToPubProto([]peer.Peer{coordPeer})

	ownerID, clusterID, err := a.core.GetIDs(nil)
	if err != nil {
		return err
	}

	// Activate all nodes.
	for num, nodePublicEndpoint := range nodePublicEndpoints {
		logToCLI("activating node %3d out of %3d nodes", num+1, len(nodePublicEndpoints))
		nodeVPNIP, err := a.core.GenerateNextIP()
		if err != nil {
			a.logger.Error("generation of vpn ips failed", zap.Error(err))
			return err
		}
		nodeVpnPubKey, err := a.activateNode(nodePublicEndpoint, nodeVPNIP, initialPeers, ownerID, clusterID)
		if err != nil {
			return err
		}
		peer := peer.Peer{
			PublicEndpoint: nodePublicEndpoint,
			VPNIP:          nodeVPNIP,
			VPNPubKey:      nodeVpnPubKey,
			Role:           role.Node,
		}
		if err := a.core.AddPeer(peer); err != nil {
			return err
		}
		if err := a.joinCluster(nodePublicEndpoint); err != nil {
			return err
		}
	}

	// Manually trigger an update operation on all nodes.
	// This may be expendable in the future, depending on whether it's acceptable that it takes
	// some seconds until the nodes get all peer data via their regular update requests.
	_, peers, err := a.core.GetPeers(0)
	if err != nil {
		return err
	}
	for _, p := range peers {
		if p.Role == role.Node {
			if err := a.triggerNodeUpdate(p.PublicEndpoint); err != nil {
				a.logger.DPanic("TriggerNodeUpdate failed", zap.Error(err))
			}
		}
	}

	return nil
}

func (a *API) activateNode(nodePublicEndpoint string, nodeVPNIP string, initialPeers []*pubproto.Peer, ownerID, clusterID []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	conn, err := a.dial(ctx, nodePublicEndpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)

	resp, err := client.ActivateAsNode(ctx, &pubproto.ActivateAsNodeRequest{
		NodeVpnIp: nodeVPNIP,
		Peers:     initialPeers,
		OwnerId:   ownerID,
		ClusterId: clusterID,
	})
	if err != nil {
		a.logger.Error("node activation failed", zap.Error(err))
		return nil, err
	}

	return resp.NodeVpnPubKey, nil
}

func (a *API) makeCoordinatorPeer() (peer.Peer, error) {
	coordinatorVPNPubKey, err := a.core.GetVPNPubKey()
	if err != nil {
		a.logger.Error("could not get key", zap.Error(err))
		return peer.Peer{}, err
	}
	coordinatorPublicIP, err := a.getPublicIPAddr()
	if err != nil {
		a.logger.Error("could not get public IP", zap.Error(err))
		return peer.Peer{}, err
	}
	return peer.Peer{
		PublicEndpoint: net.JoinHostPort(coordinatorPublicIP, endpointAVPNPort),
		VPNIP:          a.core.GetCoordinatorVPNIP(),
		VPNPubKey:      coordinatorVPNPubKey,
		Role:           role.Coordinator,
	}, err
}

func (a *API) newLogToCLIFunc(send func(string) error) logFunc {
	return func(format string, v ...interface{}) {
		if err := send(fmt.Sprintf(format, v...)); err != nil {
			a.logger.Error("logging to CLI failed", zap.Error(err))
		}
	}
}

func (a *API) joinCluster(nodePublicEndpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	// We don't verify the peer certificate here, since JoinCluster triggers a connection over VPN
	// The target of the rpc needs to already be part of the VPN to process the request, meaning it is trusted
	conn, err := a.dialNoVerify(ctx, nodePublicEndpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.JoinCluster(ctx, &pubproto.JoinClusterRequest{})

	return err
}

func (a *API) triggerNodeUpdate(nodePublicEndpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	// We don't verify the peer certificate here, since TriggerNodeUpdate triggers a connection over VPN
	// The target of the rpc needs to already be part of the VPN to process the request, meaning it is trusted
	conn, err := a.dialNoVerify(ctx, nodePublicEndpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.TriggerNodeUpdate(ctx, &pubproto.TriggerNodeUpdateRequest{})

	return err
}

type logFunc func(format string, v ...interface{})
