//go:build cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"
	"libvirt.org/go/libvirt"
)

type stubNetwork struct {
	leases      []libvirt.NetworkDHCPLease
	getLeaseErr error
}

func newStubNetwork(leases []virtwrapper.NetworkDHCPLease, getLeaseErr error) stubNetwork {
	libvirtLeases := make([]libvirt.NetworkDHCPLease, len(leases))
	for i, l := range leases {
		libvirtLeases[i] = libvirt.NetworkDHCPLease{
			IPaddr:   l.IPaddr,
			Hostname: l.Hostname,
		}
	}
	return stubNetwork{
		leases:      libvirtLeases,
		getLeaseErr: getLeaseErr,
	}
}

func (n stubNetwork) GetDHCPLeases() ([]libvirt.NetworkDHCPLease, error) {
	return n.leases, n.getLeaseErr
}

func (n stubNetwork) Free() error {
	return nil
}
