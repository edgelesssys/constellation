//go:build !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import "github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/virtwrapper"

type stubNetwork struct {
	leases      []virtwrapper.NetworkDHCPLease
	getLeaseErr error
}

func newStubNetwork(leases []virtwrapper.NetworkDHCPLease, getLeaseErr error) stubNetwork {
	return stubNetwork{
		leases:      leases,
		getLeaseErr: getLeaseErr,
	}
}

func (n stubNetwork) GetDHCPLeases() ([]virtwrapper.NetworkDHCPLease, error) {
	return n.leases, n.getLeaseErr
}

func (n stubNetwork) Free() error {
	return nil
}
