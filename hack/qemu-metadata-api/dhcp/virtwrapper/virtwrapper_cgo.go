//go:build cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package virtwrapper

import (
	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/dhcp"

	"libvirt.org/go/libvirt"
)

// Connect wraps a libvirt connection.
type Connect struct {
	conn        *libvirt.Connect
	networkName string
}

// New creates a new libvirt Connct wrapper.
func New(conn *libvirt.Connect, networkName string) *Connect {
	return &Connect{
		conn:        conn,
		networkName: networkName,
	}
}

// LookupNetworkByName looks up a network by name.
func (c *Connect) lookupNetworkByName(name string) (*libvirt.Network, error) {
	net, err := c.conn.LookupNetworkByName(name)
	if err != nil {
		return nil, err
	}
	return net, nil
}

// GetDHCPLeases returns the underlying DHCP leases.
func (c *Connect) GetDHCPLeases() ([]dhcp.NetworkDHCPLease, error) {
	net, err := c.lookupNetworkByName(c.networkName)
	if err != nil {
		return nil, err
	}
	defer net.Free()

	leases, err := net.GetDHCPLeases()
	if err != nil {
		return nil, err
	}
	ret := make([]dhcp.NetworkDHCPLease, len(leases))
	for i, l := range leases {
		ret[i] = dhcp.NetworkDHCPLease{
			IPaddr:   l.IPaddr,
			Hostname: l.Hostname,
		}
	}
	return ret, nil
}
