/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package virtwrapper

import "libvirt.org/go/libvirt"

type Connect struct {
	Conn *libvirt.Connect
}

func (c *Connect) LookupNetworkByName(name string) (*Network, error) {
	net, err := c.Conn.LookupNetworkByName(name)
	if err != nil {
		return nil, err
	}
	return &Network{Net: net}, nil
}

type Network struct {
	Net virNetwork
}

func (n *Network) GetDHCPLeases() ([]libvirt.NetworkDHCPLease, error) {
	return n.Net.GetDHCPLeases()
}

func (n *Network) Free() {
	_ = n.Net.Free()
}

type virNetwork interface {
	GetDHCPLeases() ([]libvirt.NetworkDHCPLease, error)
	Free() error
}
