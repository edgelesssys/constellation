package pubapi

import (
	"context"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
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

	// Setup SSH users on subsequent coordinator, if defined
	if len(in.SshUserKeys) != 0 {
		sshUserKeys := ssh.FromProtoSlice(in.SshUserKeys)
		if err := a.core.CreateSSHUsers(sshUserKeys); err != nil {
			return nil, status.Errorf(codes.Internal, "creating SSH users on additional coordinators: %v", err)
		}
	}
	// add one coordinator to the VPN
	if err := a.core.SetVPNIP(in.AssignedVpnIp); err != nil {
		return nil, status.Errorf(codes.Internal, "set vpn IP address: %v", err)
	}

	if err := a.core.AddPeerToVPN(peer.FromPubProto([]*pubproto.Peer{in.ActivatingCoordinatorData})[0]); err != nil {
		return nil, status.Errorf(codes.Internal, "adding initial peers to vpn: %v", err)
	}

	// run the VPN-API server
	if err := a.StartVPNAPIServer(in.AssignedVpnIp); err != nil {
		return nil, status.Errorf(codes.Internal, "start vpnAPIServer: %v", err)
	}

	a.logger.Info("retrieving k8s join information ")
	joinArgs, certKey, err := a.getk8SCoordinatorJoinArgs(ctx, in.ActivatingCoordinatorData.VpnIp, vpnAPIPort)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in getk8sJoinArgs: %v", err)
	}
	// Before we join the cluster we need to be able to communicate with ALL other control-planes
	err = a.core.UpdatePeers(peer.FromPubProto(in.Peers))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "add peers to vpn: %v", err)
	}
	a.logger.Info("about to join the k8s cluster")
	err = a.core.JoinCluster(joinArgs, certKey, role.Coordinator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

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
	if err := a.core.PersistNodeState(role.Coordinator, in.AssignedVpnIp, in.OwnerId, in.ClusterId); err != nil {
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
	for _, p := range peers {
		if p.Role == role.Node {
			if err := a.triggerNodeUpdate(p.PublicIP); err != nil {
				a.logger.Error("triggerNodeUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
		if p.Role == role.Coordinator && p.VPNIP != thisPeer.VPNIP {
			if err := a.triggerCoordinatorUpdate(context.TODO(), p.PublicIP); err != nil {
				a.logger.Error("triggerCoordinatorUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
	}

	return &pubproto.ActivateAsAdditionalCoordinatorResponse{}, nil
}

func (a *API) ActivateAdditionalCoordinator(ctx context.Context, in *pubproto.ActivateAdditionalCoordinatorRequest) (*pubproto.ActivateAdditionalCoordinatorResponse, error) {
	err := a.activateCoordinator(ctx, in.CoordinatorPublicIp, in.SshUserKeys)
	if err != nil {
		a.logger.Error("coordinator activation failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "activate new coordinator: %v", err)
	}
	return &pubproto.ActivateAdditionalCoordinatorResponse{}, nil
}

func (a *API) activateCoordinators(logToCLI logFunc, coordinatorPublicIPs []string, sshUserKeys []*pubproto.SSHUserKey) error {
	// Activate all coordinators.
	for num, coordinatorPublicIP := range coordinatorPublicIPs {
		logToCLI("Activating control-plane node %3d out of %3d ...", num+2, len(coordinatorPublicIPs)+1)
		if err := a.activateCoordinator(context.TODO(), coordinatorPublicIP, sshUserKeys); err != nil {
			return err
		}
	}
	return nil
}

func (a *API) activateCoordinator(ctx context.Context, coordinatorIP string, sshUserKeys []*pubproto.SSHUserKey) error {
	ctx, cancel := context.WithTimeout(ctx, deadlineDuration)
	defer cancel()

	if err := a.core.RequireState(state.ActivatingNodes); err != nil {
		return fmt.Errorf("coordinator is not in required state: %v", err)
	}
	assignedVPNIP, err := a.core.GetNextCoordinatorIP()
	if err != nil {
		return fmt.Errorf("requesting new coordinator vpn IP address: %v", err)
	}
	vpnIP, err := a.core.GetVPNIP()
	if err != nil {
		return fmt.Errorf("get own vpn IP address: %v", err)
	}
	thisPeer, err := a.assemblePeerStruct(vpnIP, role.Coordinator)
	if err != nil {
		return fmt.Errorf("assembling coordinator peer struct: %v", err)
	}
	ownerID, clusterID, err := a.core.GetIDs(nil)
	if err != nil {
		return fmt.Errorf("get owner and cluster ID: %v", err)
	}
	_, peers, err := a.core.GetPeers(0)
	if err != nil {
		return err
	}

	conn, err := a.dialer.Dial(ctx, net.JoinHostPort(coordinatorIP, endpointAVPNPort))
	if err != nil {
		return fmt.Errorf("dialing new coordinator: %v", err)
	}
	defer conn.Close()
	client := pubproto.NewAPIClient(conn)
	// This call can be omitted, since this function will be called by the ToActivating Coordinator
	// and he knows his own PubKey, so he can pass this as argument.
	// TODO: Remove this gRPC function when we have working integration.
	resp, err := client.GetPeerVPNPublicKey(ctx, &pubproto.GetPeerVPNPublicKeyRequest{})
	if err != nil {
		a.logger.Error("failed to get PubKey from new coordinator", zap.Error(err))
		return err
	}
	newCoordinatorPeer := peer.Peer{VPNIP: assignedVPNIP, PublicIP: coordinatorIP, VPNPubKey: resp.CoordinatorPubKey, Role: role.Coordinator}
	err = a.core.AddPeer(newCoordinatorPeer)
	if err != nil {
		a.logger.Error("failed to store new coordinator data", zap.Error(err))
		return err
	}
	for _, p := range peers {
		if p.Role == role.Coordinator && p.VPNIP != thisPeer.VPNIP {
			if err := a.triggerCoordinatorUpdate(context.TODO(), p.PublicIP); err != nil {
				a.logger.Error("triggerCoordinatorUpdate failed", zap.Error(err), zap.String("endpoint", p.PublicIP), zap.String("vpnip", p.VPNIP))
			}
		}
	}
	_, err = client.ActivateAsAdditionalCoordinator(ctx, &pubproto.ActivateAsAdditionalCoordinatorRequest{
		AssignedVpnIp:             assignedVPNIP,
		ActivatingCoordinatorData: peer.ToPubProto([]peer.Peer{thisPeer})[0],
		Peers:                     peer.ToPubProto(peers),
		OwnerId:                   ownerID,
		ClusterId:                 clusterID,
		SshUserKeys:               sshUserKeys,
	})
	return err
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

// GetPeerVPNPublicKey return the VPN publicKey of the peer.
func (a *API) GetPeerVPNPublicKey(ctx context.Context, in *pubproto.GetPeerVPNPublicKeyRequest) (*pubproto.GetPeerVPNPublicKeyResponse, error) {
	key, err := a.core.GetVPNPubKey()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not obtain VPNPubKey %v", err)
	}
	return &pubproto.GetPeerVPNPublicKeyResponse{CoordinatorPubKey: key}, nil
}

func (a *API) triggerCoordinatorUpdate(ctx context.Context, publicIP string) error {
	ctx, cancel := context.WithTimeout(ctx, deadlineDuration)
	defer cancel()

	// We don't verify the peer certificate here, since TriggerNodeUpdate triggers a connection over VPN
	// The target of the rpc needs to already be part of the VPN to process the request, meaning it is trusted
	conn, err := a.dialer.DialNoVerify(ctx, net.JoinHostPort(publicIP, endpointAVPNPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.TriggerCoordinatorUpdate(ctx, &pubproto.TriggerCoordinatorUpdateRequest{})

	return err
}

func (a *API) getk8SCoordinatorJoinArgs(ctx context.Context, coordinatorIP, port string) (*kubeadm.BootstrapTokenDiscovery, string, error) {
	conn, err := a.dialer.DialInsecure(ctx, net.JoinHostPort(coordinatorIP, port))
	if err != nil {
		return nil, "", err
	}
	defer conn.Close()
	client := vpnproto.NewAPIClient(conn)
	// since the key has to generated every time, this gRPC induces ~1s overhead.
	resp, err := client.GetK8SCertificateKey(ctx, &vpnproto.GetK8SCertificateKeyRequest{})
	if err != nil {
		return nil, "", err
	}
	joinArgs, err := client.GetK8SJoinArgs(ctx, &vpnproto.GetK8SJoinArgsRequest{})
	if err != nil {
		return nil, "", err
	}
	joinToken := &kubeadm.BootstrapTokenDiscovery{
		Token:             joinArgs.Token,
		APIServerEndpoint: joinArgs.ApiServerEndpoint,
		CACertHashes:      []string{joinArgs.DiscoveryTokenCaCertHash},
	}

	return joinToken, resp.CertificateKey, err
}
