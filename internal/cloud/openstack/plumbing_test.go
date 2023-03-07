/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServerAddresses(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected []serverSubnetAddresses
		wantErr  bool
	}{
		"valid": {
			input: `
				{
					"network1": [
						{
							"OS-EXT-IPS-MAC:mac_addr": "00:00:00:00:00:00",
							"version": 4,
							"addr": "192.0.2.1",
							"OS-EXT-IPS:type": "fixed"
						},
						{
							"OS-EXT-IPS-MAC:mac_addr": "00:00:00:00:00:01",
							"version": 6,
							"addr": "2001:db8:3333:4444:5555:6666:7777:8888",
							"OS-EXT-IPS:type": "floating"
						}
					],
					"network2": [
						{
							"OS-EXT-IPS-MAC:mac_addr": "00:00:00:00:00:02",
							"version": 4,
							"addr": "192.0.2.3",
							"OS-EXT-IPS:type": "floating"
						}
					]
				}`,
			expected: []serverSubnetAddresses{
				{
					NetworkName: "network1",
					Addresses: []serverAddress{
						{
							Type:      fixedIP,
							IPVersion: ipV4,
							IP:        "192.0.2.1",
							MAC:       "00:00:00:00:00:00",
						},
						{
							Type:      floatingIP,
							IPVersion: ipV6,
							IP:        "2001:db8:3333:4444:5555:6666:7777:8888",
							MAC:       "00:00:00:00:00:01",
						},
					},
				},
				{
					NetworkName: "network2",
					Addresses: []serverAddress{
						{
							Type:      floatingIP,
							IPVersion: ipV4,
							IP:        "192.0.2.3",
							MAC:       "00:00:00:00:00:02",
						},
					},
				},
			},
		},
		"invalid ip version": {
			input: `
				{
					"network1": [
						{
							"OS-EXT-IPS-MAC:mac_addr": "00:00:00:00:00:00",
							"version": 5,
							"addr": "192.0.2.1",
							"OS-EXT-IPS:type": "fixed"
						}
					]
				}`,
			wantErr: true,
		},
		"invalid ip type": {
			input: `
				{
					"network1": [
						{
							"OS-EXT-IPS-MAC:mac_addr": "00:00:00:00:00:00",
							"version": 4,
							"addr": "192.0.2.1",
							"OS-EXT-IPS:type": "invalid"
						}
					]
				}`,
			wantErr: true,
		},
		"invalid second lvl structure": {
			input: `
				{
					"network1": { "invalid": "structure" }
				}`,
			wantErr: true,
		},
		"invalid third lvl structure": {
			input: `
				{
					"network1": [ "invalid" ]
				}`,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var addrsMap map[string]any
			err := json.Unmarshal([]byte(tc.input), &addrsMap)
			require.NoError(err)

			addrs, err := parseSeverAddresses(addrsMap)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.ElementsMatch(tc.expected, addrs)
		})
	}
}
