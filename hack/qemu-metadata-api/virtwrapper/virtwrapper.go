/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package virtwrapper

import "libvirt.org/go/libvirt"

// Connect wraps a libvirt connection.
type Connect struct {
	Conn *libvirt.Connect
}

// LookupNetworkByName looks up a network by name.
func (c *Connect) LookupNetworkByName(name string) (*Network, error) {
	net, err := c.Conn.LookupNetworkByName(name)
	if err != nil {
		return nil, err
	}
	return &Network{Net: net}, nil
}

// Network wraps a libvirt network.
type Network struct {
	Net virNetwork
}

// GetDHCPLeases returns the underlying DHCP leases.
func (n *Network) GetDHCPLeases() ([]libvirt.NetworkDHCPLease, error) {
	return n.Net.GetDHCPLeases()
}

// Free the network resource.
func (n *Network) Free() {
	_ = n.Net.Free()
}

type virNetwork interface {
	GetDHCPLeases() ([]libvirt.NetworkDHCPLease, error)
	Free() error
}
