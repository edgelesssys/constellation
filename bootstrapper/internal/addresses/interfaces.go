/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package addresses

import (
	"net"
)

// GetMachineNetworkAddresses retrieves all network interface addresses.
func GetMachineNetworkAddresses() ([]string, error) {
	var addresses []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip.IsLoopback() {
				continue
			}
			addresses = append(addresses, ip.String())
		}
	}

	return addresses, nil
}
