package pubapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/edgelesssys/constellation/coordinator/config"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ActivateAsCoordinator is the RPC call to activate the Coordinator.
func (a *API) ActivateAsCoordinator(in *pubproto.ActivateAsCoordinatorRequest, srv pubproto.API_ActivateAsCoordinatorServer) (reterr error) {
	a.mut.Lock()
	defer a.mut.Unlock()

	a.cloudLogger.Disclose("ActivateAsCoordinator called.")

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

	logToCLI("Initializing first control-plane node ...")

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

	// Setup SSH users for the first coordinator, if defined
	if len(in.SshUserKeys) != 0 {
		logToCLI("Creating SSH users on first control-plane node...")
		sshUserKeys := ssh.FromProtoSlice(in.SshUserKeys)
		if err := a.core.CreateSSHUsers(sshUserKeys); err != nil {
			return status.Errorf(codes.Internal, "creating SSH users: %v", err)
		}
	}

	logToCLI("Initializing Kubernetes ...")
	id := attestationtypes.ID{Owner: ownerID, Cluster: clusterID}
	kubeconfig, err := a.core.InitCluster(context.TODO(), in.AutoscalingNodeGroups, in.CloudServiceAccountUri, id, in.MasterSecret, in.SshUserKeys)
	if err != nil {
		return status.Errorf(codes.Internal, "initializing Kubernetes cluster failed: %v", err)
	}

	// run the VPN-API server
	if err := a.StartVPNAPIServer(coordPeer.VPNIP); err != nil {
		return status.Errorf(codes.Internal, "start vpnAPIServer: %v", err)
	}
	if err := a.core.SwitchToPersistentStore(); err != nil {
		return status.Errorf(codes.Internal, "switch to persistent store: %v", err)
	}

	// TODO: check performance and maybe make concurrent
	if err := a.activateCoordinators(logToCLI, in.CoordinatorPublicIps, in.SshUserKeys); err != nil {
		a.logger.Error("coordinator activation failed", zap.Error(err))
		return status.Errorf(codes.Internal, "coordinator initialization: %v", err)
	}
	// TODO: check performance and maybe make concurrent
	if err := a.activateNodes(logToCLI, in.NodePublicIps, in.SshUserKeys); err != nil {
		a.logger.Error("node activation failed", zap.Error(err))
		return status.Errorf(codes.Internal, "node initialization: %v", err)
	}
	// persist node state on disk
	if err := a.core.PersistNodeState(role.Coordinator, coordPeer.VPNIP, ownerID, clusterID); err != nil {
		return status.Errorf(codes.Internal, "persist node state: %v", err)
	}
	diskUUID, err := a.core.GetDiskUUID()
	if err != nil {
		return status.Errorf(codes.Internal, "getting disk uuid: %v", err)
	}
	diskKey, err := a.core.GetDataKey(ctx, diskUUID, 32)
	if err != nil {
		return status.Errorf(codes.Internal, "getting disk key: %v", err)
	}
	if err := a.core.UpdateDiskPassphrase(string(diskKey)); err != nil {
		return status.Errorf(codes.Internal, "updating disk key: %v", err)
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
	a.cloudLogger.Disclose("ActivateAdditionalNodes called.")

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
	if err := a.activateNodes(logToCLI, in.NodePublicIps, in.SshUserKeys); err != nil {
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
	a.cloudLogger.Disclose("RequestStateDiskKey called.")
	if err := a.core.RequireState(state.ActivatingNodes); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	key, err := a.core.GetDataKey(ctx, in.DiskUuid, config.RNGLengthDefault)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to load key: %v", err)
	}

	peer, err := a.peerFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	conn, err := a.dialer.Dial(ctx, peer)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	defer conn.Close()

	client := keyproto.NewAPIClient(conn)
	if _, err := client.PushStateDiskKey(ctx, &keyproto.PushStateDiskKeyRequest{StateDiskKey: key}); err != nil {
		return nil, status.Errorf(codes.Internal, "pushing key to peer %q: %v", peer, err)
	}

	return &pubproto.RequestStateDiskKeyResponse{}, nil
}

func (a *API) activateNodes(logToCLI logFunc, nodePublicIPs []string, sshUserKeys []*pubproto.SSHUserKey) error {
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
		logToCLI("Activating worker node %3d out of %3d ...", num+1, len(nodePublicIPs))
		nodeVPNIP, err := a.core.GetNextNodeIP()
		if err != nil {
			a.logger.Error("generation of vpn ips failed", zap.Error(err))
			return err
		}
		nodeVpnPubKey, err := a.activateNode(nodePublicIP, nodeVPNIP, initialPeers, ownerID, clusterID, sshUserKeys)
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

func (a *API) activateNode(nodePublicIP string, nodeVPNIP string, initialPeers []*pubproto.Peer, ownerID, clusterID []byte, sshUserKeys []*pubproto.SSHUserKey) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), deadlineDuration)
	defer cancel()

	conn, err := a.dialer.Dial(ctx, net.JoinHostPort(nodePublicIP, endpointAVPNPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)

	stream, err := client.ActivateAsNode(ctx)
	if err != nil {
		a.logger.Error("connecting to node for activation failed", zap.Error(err))
		return nil, err
	}

	/*
		coordinator -> initial request -> node
	*/
	if err := stream.Send(&pubproto.ActivateAsNodeRequest{
		Request: &pubproto.ActivateAsNodeRequest_InitialRequest{
			InitialRequest: &pubproto.ActivateAsNodeInitialRequest{
				NodeVpnIp:   nodeVPNIP,
				Peers:       initialPeers,
				OwnerId:     ownerID,
				ClusterId:   clusterID,
				SshUserKeys: sshUserKeys,
			},
		},
	}); err != nil {
		a.logger.Error("sending initial message to node for activation failed", zap.Error(err))
		return nil, err
	}

	/*
		coordinator <- state disk uuid <- node
	*/
	// wait for message containing the nodes disk UUID to send back the permanent encryption key
	message, err := stream.Recv()
	if err != nil {
		a.logger.Error("expected disk UUID message but no message received", zap.Error(err))
		return nil, err
	}
	diskUUID, ok := message.GetResponse().(*pubproto.ActivateAsNodeResponse_StateDiskUuid)
	if !ok {
		a.logger.Error("expected disk UUID message but got different message")
		return nil, errors.New("expected state disk UUID but got different message type")
	}
	diskKey, err := a.core.GetDataKey(ctx, diskUUID.StateDiskUuid, 32)
	if err != nil {
		a.logger.Error("failed to derive node's disk key")
		return nil, err
	}

	/*
		coordinator -> state disk key -> node
	*/
	// send back state disk encryption key
	if err := stream.Send(&pubproto.ActivateAsNodeRequest{
		Request: &pubproto.ActivateAsNodeRequest_StateDiskKey{
			StateDiskKey: diskKey,
		},
	}); err != nil {
		a.logger.Error("sending state disk key to node on activation failed", zap.Error(err))
		return nil, err
	}

	/*
		coordinator <- VPN public key <- node
	*/
	// wait for message containing the node VPN pubkey
	message, err = stream.Recv()
	if err != nil {
		a.logger.Error("expected node VPN pubkey but no message received", zap.Error(err))
		return nil, err
	}
	vpnPubKey, ok := message.GetResponse().(*pubproto.ActivateAsNodeResponse_NodeVpnPubKey)
	if !ok {
		a.logger.Error("expected node VPN pubkey but got different message")
		return nil, errors.New("expected node VPN pub key but got different message type")
	}

	return vpnPubKey.NodeVpnPubKey, nil
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
	return func(format string, v ...any) {
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
	conn, err := a.dialer.DialNoVerify(ctx, net.JoinHostPort(nodePublicIP, endpointAVPNPort))
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
	conn, err := a.dialer.DialNoVerify(ctx, net.JoinHostPort(nodePublicIP, endpointAVPNPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.TriggerNodeUpdate(ctx, &pubproto.TriggerNodeUpdateRequest{})

	return err
}

type logFunc func(format string, v ...any)
