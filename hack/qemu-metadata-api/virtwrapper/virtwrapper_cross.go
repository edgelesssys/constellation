//go:build !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package virtwrapper

import "errors"

// Connect wraps a libvirt connection.
type Connect struct{}

// LookupNetworkByName looks up a network by name.
// This function errors if CGO is disabled.
func (c *Connect) LookupNetworkByName(_ string) (*Network, error) {
	return nil, errors.New("using virtwrapper requires building with CGO")
}

// Network wraps a libvirt network.
type Network struct {
	Net Net
}

// GetDHCPLeases returns the underlying DHCP leases.
// This function errors if CGO is disabled.
func (n *Network) GetDHCPLeases() ([]NetworkDHCPLease, error) {
	return n.Net.GetDHCPLeases()
}

// Free the network resource.
// This function does nothing if CGO is disabled.
func (n *Network) Free() {}

// Net is a libvirt Network.
type Net interface {
	GetDHCPLeases() ([]NetworkDHCPLease, error)
}
