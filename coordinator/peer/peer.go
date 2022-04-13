package peer

import (
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
)

// Peer holds all information about a peer.
type Peer struct {
	// PublicIP is the public IP address on which the peer is reachable.
	PublicIP string
	// VPNIP holds the internal VPN address, can only be used within the VPN
	// and some gRPC services may only be reachable from this IP.
	VPNIP string
	// VPNPubKey contains the PublicKey used for cryptographic purposes in the VPN.
	VPNPubKey []byte
	// Role is the peer's role (Coordinator, Node or Admin).
	Role role.Role
}

func FromPubProto(peers []*pubproto.Peer) []Peer {
	var result []Peer
	for _, p := range peers {
		result = append(result, Peer{
			PublicIP:  p.PublicIp,
			VPNIP:     p.VpnIp,
			VPNPubKey: p.VpnPubKey,
			Role:      role.Role(p.Role),
		})
	}
	return result
}

func ToPubProto(peers []Peer) []*pubproto.Peer {
	var result []*pubproto.Peer
	for _, p := range peers {
		result = append(result, &pubproto.Peer{
			PublicIp:  p.PublicIP,
			VpnIp:     p.VPNIP,
			VpnPubKey: p.VPNPubKey,
			Role:      uint32(p.Role),
		})
	}
	return result
}

func FromVPNProto(peers []*vpnproto.Peer) []Peer {
	var result []Peer
	for _, p := range peers {
		result = append(result, Peer{
			PublicIP:  p.PublicIp,
			VPNIP:     p.VpnIp,
			VPNPubKey: p.VpnPubKey,
			Role:      role.Role(p.Role),
		})
	}
	return result
}

func ToVPNProto(peers []Peer) []*vpnproto.Peer {
	var result []*vpnproto.Peer
	for _, p := range peers {
		result = append(result, &vpnproto.Peer{
			PublicIp:  p.PublicIP,
			VpnIp:     p.VPNIP,
			VpnPubKey: p.VPNPubKey,
			Role:      uint32(p.Role),
		})
	}
	return result
}
