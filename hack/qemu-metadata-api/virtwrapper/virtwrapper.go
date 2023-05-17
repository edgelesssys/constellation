/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package virtwrapper

// NetworkDHCPLease abstracts a libvirt DHCP lease.
type NetworkDHCPLease struct {
	IPaddr   string
	Hostname string
}
