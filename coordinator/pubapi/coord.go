package pubapi

import (
	"context"
	"errors"
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
		return status.Errorf(codes.FailedPrecondition, "node is not in required state: %v", err)
	}

	if len(in.MasterSecret) == 0 {
		a.logger.Error("missing master secret")
		return status.Error(codes.InvalidArgument, "missing master secret")
	}

	logToCLI := a.newLogToCLIFunc(func(msg string) error {
		return srv.Send(&pubproto.ActivateAsCoordinatorResponse{
			Content: &pubproto.ActivateAsCoordinatorResponse_Log{
				Log: &pubproto.Log{
					Message: msg,
				},
			},
		})
	})

	logToCLI("Initializing first Coordinator ...")

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
	if err := a.core.InitializeStoreIPs(); err != nil {
		return status.Errorf(codes.Internal, "initialize store IPs: %v", err)
	}

	ownerID, clusterID, err := a.core.GetIDs(in.MasterSecret)
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}
	if err := a.core.AdvanceState(state.ActivatingNodes, ownerID, clusterID); err != nil {
		return status.Errorf(codes.Internal, "advance state to ActivatingNodes: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := a.core.SetUpKMS(ctx, in.StorageUri, in.KmsUri, in.KeyEncryptionKeyId, in.UseExistingKek); err != nil {
		return status.Errorf(codes.Internal, "setting up KMS: %v", err)
	}
	vpnIP, err := a.core.GetNextCoordinatorIP()
	if err != nil {
		return status.Errorf(codes.Internal, "get coordinator vpn IP address: %v", err)
	}
	coordPeer, err := a.assemblePeerStruct(vpnIP, role.Coordinator)
	if err != nil {
		return status.Errorf(codes.Internal, "assembling the coordinator peer struct: %v", err)
	}

	if err := a.core.SetVPNIP(coordPeer.VPNIP); err != nil {
		return status.Errorf(codes.Internal, "set the vpn IP address: %v", err)
	}
	if err := a.core.AddPeer(coordPeer); err != nil {
		return status.Errorf(codes.Internal, "adding the coordinator to store/vpn: %v", err)
	}

	logToCLI("Initializing Kubernetes ...")
	kubeconfig, err := a.core.InitCluster(in.AutoscalingNodeGroups, in.CloudServiceAccountUri)
	if err != nil {
		return status.Errorf(codes.Internal, "initializing kubernetes cluster failed: %v", err)
	}

	// run the VPN-API server
	if err := a.vpnAPIServer.Listen(net.JoinHostPort(coordPeer.VPNIP, vpnAPIPort)); err != nil {
		return status.Errorf(codes.Internal, "start vpnAPIServer: %v", err)
	}
	a.wgClose.Add(1)
	go func() {
		defer a.wgClose.Done()
		if err := a.vpnAPIServer.Serve(); err != nil {
			panic(err)
		}
	}()
	// TODO: check performance and maybe make concurrent
	if err := a.activateNodes(logToCLI, in.NodePublicIps); err != nil {
		a.logger.Error("node activation failed", zap.Error(err))
		return status.Errorf(codes.Internal, "node initialization: %v", err)
	}

	if err := a.core.SwitchToPersistentStore(); err != nil {
		return status.Errorf(codes.Internal, "switch to persistent store: %v", err)
	}

	// persist node state on disk
	if err := a.core.PersistNodeState(role.Coordinator, ownerID, clusterID); err != nil {
		return status.Errorf(codes.Internal, "persist node state: %v", err)
	}
	adminVPNIP, err := a.core.GetNextNodeIP()
	if err != nil {
		return status.Errorf(codes.Internal, "requesting node IP address: %v", err)
	}
	// This effectively gives code execution, so we do this last.
	err = a.core.AddPeer(peer.Peer{
		VPNIP:     adminVPNIP,
		VPNPubKey: in.AdminVpnPubKey,
		Role:      role.Admin,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "add peer to store/vpn: %v", err)
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

	logToCLI := a.newLogToCLIFunc(func(msg string) error {
		return srv.Send(&pubproto.ActivateAdditionalNodesResponse{
			Log: &pubproto.Log{
				Message: msg,
			},
		})
	})

	// TODO: check performance and maybe make concurrent
	if err := a.activateNodes(logToCLI, in.NodePublicIps); err != nil {
		a.logger.Error("node activation failed", zap.Error(err))
		return status.Errorf(codes.Internal, "activating nodes: %v", err)
	}

	return srv.Send(&pubproto.ActivateAdditionalNodesResponse{
		Log: &pubproto.Log{
			Message: "success",
		},
	})
}

// RequestStateDiskKey triggers the Coordinator to return a key derived from the Constellation's master secret to the caller.
func (a *API) RequestStateDiskKey(ctx context.Context, in *pubproto.RequestStateDiskKeyRequest) (*pubproto.RequestStateDiskKeyResponse, error) {
	// TODO: Add Coordinator call to restarting node and deliver the key
	/*
		if err := a.core.RequireState(state.IsNode, state.ActivatingNodes); err != nil {
			return nil, err
		}
		_, err := a.core.GetDataKey(ctx, in.DiskUuid, 32)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	*/
	return &pubproto.RequestStateDiskKeyResponse{}, errors.New("unimplemented")
}

func (a *API) activateNodes(logToCLI logFunc, nodePublicIPs []string) error {
	_, peers, err := a.core.GetPeers(0)
	if err != nil {
		return err
	}
	// we need to add at least all coordinators to the peer for HA
	initialPeers := peer.ToPubProto(peers)

	ownerID, clusterID, err := a.core.GetIDs(nil)
	if err != nil {
		return err
	}

	// Activate all nodes.
	for num, nodePublicIP := range nodePublicIPs {
		logToCLI("activating node %3d out of %3d nodes ...", num+1, len(nodePublicIPs))
		nodeVPNIP, err := a.core.GetNextNodeIP()
		if err != nil {
			a.logger.Error("generation of vpn ips failed", zap.Error(err))
			return err
		}
		nodeVpnPubKey, err := a.activateNode(nodePublicIP, nodeVPNIP, initialPeers, ownerID, clusterID)
		if err != nil {
			return err
		}
		peer := peer.Peer{
			PublicIP:  nodePublicIP,
			VPNIP:     nodeVPNIP,
			VPNPubKey: nodeVpnPubKey,
			Role:      role.Node,
		}
		if err := a.core.AddPeer(peer); err != nil {
			return err
		}
		// This can be omitted if we
		// 1. Use a gRPC HA balancer mechanism, which picks one active coordinator connection
		// (nodeUpdate loop causes problems, even if we specify the IP in the joinCluster RPC)
		if err := a.updateCoordinator(); err != nil {
			return err
		}
		if err := a.joinCluster(nodePublicIP); err != nil {
			return err
		}
	}

	// Manually trigger an update operation on all peers.
	// This may be expendable in the future, depending on whether it's acceptable that it takes
	// some seconds until the nodes get all peer data via their regular update requests.
	_, peers, err = a.core.GetPeers(0)
	if err != nil {
		return err
	}
	vpnIP, err := a.core.GetVPNIP()
	if err != nil {
		return err
	}
	for _, p := range peers {
		if p.Role == role.Node {
			if err := a.triggerNodeUpdate(p.PublicIP); err != nil {
				a.logger.Error("TriggerNodeUpdate failed", zap.Error(err))
			}
		}

		if p.Role == role.Coordinator && p.VPNIP != vpnIP {
			a.logger.Info("update coordinator", zap.String("coordinator vpnIP", p.VPNIP))
			if err := a.triggerCoordinatorUpdate(context.TODO(), p.PublicIP); err != nil {
				// no reason to panic here, we can recover
				a.logger.Error("triggerCoordinatorUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
	}

	return nil
}

func (a *API) activateNode(nodePublicIP string, nodeVPNIP string, initialPeers []*pubproto.Peer, ownerID, clusterID []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	conn, err := a.dial(ctx, net.JoinHostPort(nodePublicIP, endpointAVPNPort))
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

// assemblePeerStruct combines all information of this peer into a peer struct.
func (a *API) assemblePeerStruct(vpnIP string, _ role.Role) (peer.Peer, error) {
	vpnPubKey, err := a.core.GetVPNPubKey()
	if err != nil {
		a.logger.Error("could not get key", zap.Error(err))
		return peer.Peer{}, err
	}
	publicIP, err := a.getPublicIPAddr()
	if err != nil {
		a.logger.Error("could not get public IP", zap.Error(err))
		return peer.Peer{}, err
	}
	return peer.Peer{
		PublicIP:  publicIP,
		VPNIP:     vpnIP,
		VPNPubKey: vpnPubKey,
		Role:      role.Coordinator,
	}, err
}

func (a *API) newLogToCLIFunc(send func(string) error) logFunc {
	return func(format string, v ...interface{}) {
		if err := send(fmt.Sprintf(format, v...)); err != nil {
			a.logger.Error("logging to CLI failed", zap.Error(err))
		}
	}
}

func (a *API) joinCluster(nodePublicIP string) error {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	vpnIP, err := a.core.GetVPNIP()
	if err != nil {
		return err
	}
	// We don't verify the peer certificate here, since JoinCluster triggers a connection over VPN
	// The target of the rpc needs to already be part of the VPN to process the request, meaning it is trusted
	conn, err := a.dialNoVerify(ctx, net.JoinHostPort(nodePublicIP, endpointAVPNPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.JoinCluster(ctx, &pubproto.JoinClusterRequest{CoordinatorVpnIp: vpnIP})

	return err
}

func (a *API) updateCoordinator() error {
	_, peers, err := a.core.GetPeers(0)
	if err != nil {
		return err
	}
	vpnIP, err := a.core.GetVPNIP()
	if err != nil {
		return err
	}
	for _, p := range peers {
		if p.Role == role.Coordinator && p.VPNIP != vpnIP {
			a.logger.Info("update coordinator", zap.String("coordinator vpnIP", p.VPNIP))
			if err := a.triggerCoordinatorUpdate(context.TODO(), p.PublicIP); err != nil {
				a.logger.Error("triggerCoordinatorUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
	}
	return nil
}

func (a *API) triggerNodeUpdate(nodePublicIP string) error {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	// We don't verify the peer certificate here, since TriggerNodeUpdate triggers a connection over VPN
	// The target of the rpc needs to already be part of the VPN to process the request, meaning it is trusted
	conn, err := a.dialNoVerify(ctx, net.JoinHostPort(nodePublicIP, endpointAVPNPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.TriggerNodeUpdate(ctx, &pubproto.TriggerNodeUpdateRequest{})

	return err
}

type logFunc func(format string, v ...interface{})
