package core

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"go.uber.org/zap"
)

const (
	callTimeout                         = 20 * time.Second
	retrieveInitialVPNPeersRetryBackoff = 60 * time.Second
)

// ReinitializeAsCoordinator re-initializes a coordinator.
func (c *Core) ReinitializeAsCoordinator(ctx context.Context, dialer Dialer, vpnIP string, api PubAPI, retryBackoff time.Duration) error {
	if err := c.SetVPNIP(vpnIP); err != nil {
		return fmt.Errorf("set vpn IP address: %v", err)
	}

	// TODO: implement (manual) recovery endpoint in cases where no other coordinators are available
	// or when etcd quorum is lost (when leader election fails)

	ownCoordinatorEndpoint := net.JoinHostPort(vpnIP, coordinatorPort)
	// try to find active coordinator to add as initial VPN peer
	// retry until coordinator is found
	var (
		initialVPNPeers []peer.Peer
		err             error
	)
	for {
		initialVPNPeers, err = c.initialVPNPeersRetriever(ctx, dialer, c.zaplogger, c.metadata, &ownCoordinatorEndpoint)
		if err == nil {
			break
		}
		time.Sleep(retryBackoff)
	}

	// add initial peers to the VPN
	if err := c.UpdatePeers(initialVPNPeers); err != nil {
		return fmt.Errorf("adding initial peers to vpn: %v", err)
	}

	// run the VPN-API server
	if err := api.StartVPNAPIServer(vpnIP); err != nil {
		return fmt.Errorf("start vpnAPIServer: %v", err)
	}

	// ATTENTION: STORE HAS TO BE EMPTY (NO OVERLAPPING KEYS) WHEN THIS FUNCTION IS CALLED
	if err := c.SwitchToPersistentStore(); err != nil {
		return fmt.Errorf("switch to persistent store: %v", err)
	}
	c.zaplogger.Info("Transition to persistent store successful")

	kmsData, err := c.GetKMSInfo()
	if err != nil {
		return fmt.Errorf("get kms info: %v", err)
	}
	if err := c.SetUpKMS(ctx, kmsData.StorageUri, kmsData.KmsUri, kmsData.KeyEncryptionKeyID, false); err != nil {
		return fmt.Errorf("setup kms: %v", err)
	}
	return nil
}

// ReinitializeAsNode re-initializes a node.
func (c *Core) ReinitializeAsNode(ctx context.Context, dialer Dialer, vpnIP string, api PubAPI, retryBackoff time.Duration) error {
	if err := c.SetVPNIP(vpnIP); err != nil {
		return fmt.Errorf("set vpn IP address: %v", err)
	}

	// try to find active coordinator to add as initial VPN peer
	// retry until coordinator is found
	var (
		initialVPNPeers []peer.Peer
		err             error
	)
	for {
		initialVPNPeers, err = c.initialVPNPeersRetriever(ctx, dialer, c.zaplogger, c.metadata, nil)
		if err == nil {
			break
		}
		time.Sleep(retryBackoff)
	}

	// add initial peers to the VPN
	if err := c.UpdatePeers(initialVPNPeers); err != nil {
		return fmt.Errorf("adding initial peers to vpn: %v", err)
	}

	api.StartUpdateLoop()
	return nil
}

func getInitialVPNPeers(ctx context.Context, dialer Dialer, logger *zap.Logger, metadata ProviderMetadata, ownCoordinatorEndpoint *string) ([]peer.Peer, error) {
	coordinatorEndpoints, err := CoordinatorEndpoints(ctx, metadata)
	if err != nil {
		return nil, fmt.Errorf("get coordinator endpoints: %v", err)
	}
	// shuffle endpoints using PRNG. While this is not a cryptographically secure random seed,
	// it is good enough for loadbalancing.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(coordinatorEndpoints), func(i, j int) {
		coordinatorEndpoints[i], coordinatorEndpoints[j] = coordinatorEndpoints[j], coordinatorEndpoints[i]
	})

	// try to find active coordinator to retrieve peers
	for _, coordinatorEndpoint := range coordinatorEndpoints {
		if ownCoordinatorEndpoint != nil && coordinatorEndpoint == *ownCoordinatorEndpoint {
			continue
		}
		callCTX, cancel := context.WithTimeout(ctx, callTimeout)
		defer cancel()
		conn, err := dialer.Dial(callCTX, coordinatorEndpoint)
		if err != nil {
			logger.Warn("getting VPN peer information from coordinator failed: dialing failed: ", zap.String("endpoint", coordinatorEndpoint), zap.Error(err))
			continue
		}
		defer conn.Close()
		client := pubproto.NewAPIClient(conn)
		resp, err := client.GetVPNPeers(ctx, &pubproto.GetVPNPeersRequest{})
		if err != nil {
			logger.Warn("getting VPN peer information from coordinator failed: request failed: ", zap.String("endpoint", coordinatorEndpoint), zap.Error(err))
			continue
		}
		return peer.FromPubProto(resp.Peers), nil
	}

	return nil, fmt.Errorf("no active coordinator found. tried %v", coordinatorEndpoints)
}

// PubAPI is the interface for the public API of the coordinator.
type PubAPI interface {
	StartVPNAPIServer(vpnIP string) error
	StartUpdateLoop()
}

type initialVPNPeersRetriever func(ctx context.Context, dialer Dialer, logger *zap.Logger, metadata ProviderMetadata, ownCoordinatorEndpoint *string) ([]peer.Peer, error)
