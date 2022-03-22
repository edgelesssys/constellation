package core

import (
	"net"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/storewrapper"
	"go.uber.org/multierr"
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
			c.zaplogger.Error("Failed to add peer to VPN", zap.Error(err), zap.String("public_ip", publicIP), zap.String("vpn_ip", peer.VPNIP))
			return err
		}
		c.zaplogger.Info("Added peer to VPN", zap.String("public_ip", publicIP), zap.String("vpn_ip", peer.VPNIP))
	}

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

	return tx.Commit()
}

// UpdatePeers synchronizes the peers known to the store and the vpn with the passed peers.
func (c *Core) UpdatePeers(peers []peer.Peer) error {
	// exclude myself
	myIP, err := c.vpn.GetInterfaceIP()
	if err != nil {
		return err
	}
	for i, p := range peers {
		if p.VPNIP == myIP {
			peers = append(peers[:i], peers[i+1:]...)
			break
		}
	}

	tx, err := c.store.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	added, removed, err := storewrapper.StoreWrapper{Store: tx}.UpdatePeers(peers)
	if err != nil {
		return err
	}

	// perform remove and add on the vpn
	var vpnErr error
	for _, p := range removed {
		vpnErr = multierr.Append(vpnErr, c.vpn.RemovePeer(p.VPNPubKey))
	}
	for _, p := range added {
		publicIP, _, err := net.SplitHostPort(p.PublicEndpoint)
		if err != nil {
			vpnErr = multierr.Append(vpnErr, err)
			continue
		}
		vpnErr = multierr.Append(vpnErr, c.vpn.AddPeer(p.VPNPubKey, publicIP, p.VPNIP))
	}
	if vpnErr != nil {
		return vpnErr
	}

	return tx.Commit()
}
