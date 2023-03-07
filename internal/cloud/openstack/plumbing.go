/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import "fmt"

// parseServerAddresses parses the untyped Addresses field of a sever struct.
func parseSeverAddresses(addrsMap map[string]any) ([]serverSubnetAddresses, error) {
	result := []serverSubnetAddresses{}
	for name, v := range addrsMap {
		subnet := serverSubnetAddresses{NetworkName: name}

		v, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("server address %q is not of form []any", name)
		}

		for _, v := range v {
			nic := serverAddress{}

			v, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("element in server address %q is not of form []map[string]any", name)
			}

			if mac, ok := v["OS-EXT-IPS-MAC:mac_addr"].(string); ok {
				nic.MAC = mac
			}

			if ip, ok := v["addr"].(string); ok {
				nic.IP = ip
			}

			if ipVersion, ok := v["version"].(float64); ok {
				version, err := ipVersionFromFloat(ipVersion)
				if err != nil {
					return nil, err
				}
				nic.IPVersion = version
			}

			if typ, ok := v["OS-EXT-IPS:type"].(string); ok {
				ipType, err := ipTypeFromString(typ)
				if err != nil {
					return nil, err
				}
				nic.Type = ipType
			}

			subnet.Addresses = append(subnet.Addresses, nic)
		}

		result = append(result, subnet)
	}

	return result, nil
}

type serverSubnetAddresses struct {
	NetworkName string
	Addresses   []serverAddress
}

type serverAddress struct {
	Type      ipType
	IPVersion ipVersion
	IP        string
	MAC       string
}

type ipVersion int

const (
	ipVUnknown ipVersion = 0
	ipV4       ipVersion = 4
	ipV6       ipVersion = 6
)

func ipVersionFromFloat(f float64) (ipVersion, error) {
	switch f {
	case 4:
		return ipV4, nil
	case 6:
		return ipV6, nil
	default:
		return 0, fmt.Errorf("unknown IP version %f", f)
	}
}

type ipType string

const (
	unknownIP  ipType = "unknown"
	fixedIP    ipType = "fixed"
	floatingIP ipType = "floating"
)

func ipTypeFromString(s string) (ipType, error) {
	switch s {
	case "fixed":
		return fixedIP, nil
	case "floating":
		return floatingIP, nil
	default:
		return unknownIP, fmt.Errorf("unknown IP type %s", s)
	}
}
