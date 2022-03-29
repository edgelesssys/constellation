package core

import (
	"net"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/storewrapper"
	"go.uber.org/zap"
)

// GetPeers returns the stored peers if the requested version differs from the stored version.
func (c *Core) GetPeers(resourceVersion int) (int, []peer.Peer, error) {
	// Most often there's nothing to do, so first check without an expensive transaction.
	curVer, err := c.data().GetPeersResourceVersion()
	if err != nil {
		return 0, nil, err
	}
	if curVer == resourceVersion {
		return curVer, nil, nil
	}

	tx, err := c.store.BeginTransaction()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()
	txdata := storewrapper.StoreWrapper{Store: tx}

	txVer, err := txdata.GetPeersResourceVersion()
	if err != nil {
		return 0, nil, err
	}
	peers, err := txdata.GetPeers()
	if err != nil {
		return 0, nil, err
	}
	return txVer, peers, nil
}

// AddPeer adds a peer to the store and the VPN.
func (c *Core) AddPeer(peer peer.Peer) error {
	if err := c.AddPeerToVPN(peer); err != nil {
		return err
	}
	return c.AddPeerToStore(peer)
}

// AddPeerToVPN adds a peer to the the VPN.
func (c *Core) AddPeerToVPN(peer peer.Peer) error {
	publicIP, _, err := net.SplitHostPort(peer.PublicEndpoint)
	if err != nil {
		c.zaplogger.Info("SplitHostPort", zap.Error(err))
		return err
	}

	// don't add myself to vpn
	myIP, err := c.vpn.GetInterfaceIP()
	if err != nil {
		return err
	}
	if myIP != peer.VPNIP {
		if err := c.vpn.AddPeer(peer.VPNPubKey, publicIP, peer.VPNIP); err != nil {
			c.zaplogger.Error("failed to add peer to VPN", zap.Error(err), zap.String("peer public_ip", publicIP), zap.String("peer vpn_ip", peer.VPNIP))
			return err
		}
		c.zaplogger.Info("added peer to VPN", zap.String("peer public_ip", publicIP), zap.String("peer vpn_ip", peer.VPNIP))
	}
	return nil
}

// AddPeerToStore adds a peer to the store.
func (c *Core) AddPeerToStore(peer peer.Peer) error {
	tx, err := c.store.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txdata := storewrapper.StoreWrapper{Store: tx}

	if err := txdata.IncrementPeersResourceVersion(); err != nil {
		return err
	}
	if err := txdata.PutPeer(peer); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.zaplogger.Info("added peer to store", zap.String("peer public_endpoint", peer.PublicEndpoint), zap.String("peer vpn_ip", peer.VPNIP))
	return nil
}

// UpdatePeers synchronizes the peers known to the store and the vpn with the passed peers.
func (c *Core) UpdatePeers(peers []peer.Peer) error {
	return c.vpn.UpdatePeers(peers)
}
