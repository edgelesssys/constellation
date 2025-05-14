/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package addresses_test

import (
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/addresses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMachineNetworkAddresses(t *testing.T) {
	_, someAddr, err := net.ParseCIDR("10.9.0.1/24")
	require.NoError(t, err)

	testCases := map[string]struct {
		interfaces []addresses.NetInterface
		wantErr    bool
	}{
		"successful": {
			interfaces: []addresses.NetInterface{
				&mockNetInterface{
					addrs: []net.Addr{
						someAddr,
					},
				},
			},
		},
		"unsuccessful": {
			interfaces: []addresses.NetInterface{
				&mockNetInterface{addrs: nil, err: errors.New("someError")},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			addrs, err := addresses.GetMachineNetworkAddresses(tc.interfaces)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Equal([]string{"10.9.0.0"}, addrs)
				assert.NoError(err)
			}
		})
	}
}

type mockNetInterface struct {
	addrs []net.Addr
	err   error
}

func (m *mockNetInterface) Addrs() ([]net.Addr, error) {
	return m.addrs, m.err
}
