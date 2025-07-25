/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package addresses

import (
	"net"
)

// GetMachineNetworkAddresses retrieves all network interface addresses.
func GetMachineNetworkAddresses(interfaces []NetInterface) ([]string, error) {
	var addresses []string

	for _, i := range interfaces {
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

// NetInterface represents a network interface used to get network addresses.
type NetInterface interface {
	Addrs() ([]net.Addr, error)
}
