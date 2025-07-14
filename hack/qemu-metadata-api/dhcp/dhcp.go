/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package dhcp

// NetworkDHCPLease abstracts a DHCP lease.
type NetworkDHCPLease struct {
	IPaddr   string
	Hostname string
}
