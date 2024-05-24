//go:build !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package virtwrapper

import (
	"errors"

	"github.com/edgelesssys/constellation/v2/hack/qemu-metadata-api/dhcp"
)

// Connect wraps a libvirt connection.
type Connect struct{}

// GetDHCPLeases returns the underlying DHCP leases.
// This function errors if CGO is disabled.
func (n *Connect) GetDHCPLeases() ([]dhcp.NetworkDHCPLease, error) {
	return nil, errors.New("using virtwrapper requires building with CGO")
}
