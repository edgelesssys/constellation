/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package dhcp

// NetworkDHCPLease abstracts a DHCP lease.
type NetworkDHCPLease struct {
	IPaddr   string
	Hostname string
}
