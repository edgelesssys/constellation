/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/stretchr/testify/assert"
)

func TestSelf(t *testing.T) {
	someErr := fmt.Errorf("failed")

	testCases := map[string]struct {
		imds    imdsAPI
		want    metadata.InstanceMetadata
		wantErr bool
	}{
		"success": {
			imds: &stubIMDSClient{
				nameResult:       "name",
				providerIDResult: "providerID",
				roleResult:       role.ControlPlane,
				vpcIPResult:      "192.0.2.1",
			},
			want: metadata.InstanceMetadata{
				Name:       "name",
				ProviderID: "providerID",
				Role:       role.ControlPlane,
				VPCIP:      "192.0.2.1",
			},
		},
		"fail to get name": {
			imds: &stubIMDSClient{
				nameErr:          someErr,
				providerIDResult: "providerID",
				roleResult:       role.ControlPlane,
				vpcIPResult:      "192.0.2.1",
			},
			wantErr: true,
		},
		"fail to get provider ID": {
			imds: &stubIMDSClient{
				nameResult:    "name",
				providerIDErr: someErr,
				roleResult:    role.ControlPlane,
				vpcIPResult:   "192.0.2.1",
			},
			wantErr: true,
		},
		"fail to get role": {
			imds: &stubIMDSClient{
				nameResult:       "name",
				providerIDResult: "providerID",
				roleErr:          someErr,
				vpcIPResult:      "192.0.2.1",
			},
			wantErr: true,
		},
		"fail to get VPC IP": {
			imds: &stubIMDSClient{
				nameResult:       "name",
				providerIDResult: "providerID",
				roleResult:       role.ControlPlane,
				vpcIPErr:         someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &MetadataClient{imds: tc.imds}

			got, err := c.Self(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

func TestList(t *testing.T) {
	someErr := fmt.Errorf("failed")

	// newTestAddrs returns a set of raw server addresses as we would get from
	// a ListServers call and as expected by the parseSeverAddresses function.
	// The hardcoded addresses don't match what we are looking for. A valid
	// address can be injected. You can pass a second valid address to test
	// that the first valid one is chosen.
	newTestAddrs := func(vpcIP1, vpcIP2 string) map[string]any {
		return map[string]any{
			"network1": []any{
				map[string]any{
					"addr":                    "198.51.100.0",
					"version":                 4,
					"OS-EXT-IPS:type":         "fixed",
					"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:0c:0c:0c",
				},
			},
			"network2": []any{
				map[string]any{
					"addr":                    "192.0.2.1",
					"version":                 4,
					"OS-EXT-IPS:type":         "floating",
					"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:0c:0c:0c",
				},
				map[string]any{
					"addr":                    "2001:db8:3333:4444:5555:6666:7777:8888",
					"version":                 6,
					"OS-EXT-IPS:type":         "fixed",
					"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:0c:0c:0c",
				},
				map[string]any{
					"addr":                    vpcIP1,
					"version":                 4,
					"OS-EXT-IPS:type":         "fixed",
					"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:0c:0c:0c",
				},
				map[string]any{
					"addr":                    vpcIP2,
					"version":                 4,
					"OS-EXT-IPS:type":         "fixed",
					"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:0c:0c:0c",
				},
			},
		}
	}

	testCases := map[string]struct {
		imds    imdsAPI
		api     serversAPI
		want    []metadata.InstanceMetadata
		wantErr bool
	}{
		"success": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						ID:        "id1",
						Tags:      &[]string{"constellation-role-control-plane", "constellation-uid-7777"},
						Addresses: newTestAddrs("192.0.2.5", ""),
					},
					{
						Name:      "name2",
						ID:        "id2",
						Tags:      &[]string{"constellation-role-worker", "constellation-uid-7777"},
						Addresses: newTestAddrs("192.0.2.6", "192.0.2.99"),
					},
					{
						Name:      "name3",
						ID:        "id3",
						Tags:      &[]string{"constellation-role-worker", "constellation-uid-8888"},
						Addresses: newTestAddrs("198.51.100.1", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			want: []metadata.InstanceMetadata{
				{
					Name:       "name1",
					ProviderID: "id1",
					Role:       role.ControlPlane,
					VPCIP:      "192.0.2.5",
				},
				{
					Name:       "name2",
					ProviderID: "id2",
					Role:       role.Worker,
					VPCIP:      "192.0.2.6",
				},
			},
		},
		"no servers found": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"imds uid error": {
			imds:    &stubIMDSClient{uidErr: someErr},
			wantErr: true,
		},
		"list nets errors": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				netsPager: newNetPager([]networks.Network{{Name: "mynet"}}, someErr),
			},
			wantErr: true,
		},
		"list subnets error": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: stubPager{allPagesErr: someErr},
			},
			wantErr: true,
		},
		"extract subnets error": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{}, someErr),
			},
			wantErr: true,
		},
		"subnet name mismatch error": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				netsPager: newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{
					{Name: "othernet", CIDR: "192.0.2.0/24"},
					{Name: "yetanothernet", CIDR: "198.51.100.0/24"},
				}, nil),
			},
			wantErr: true,
		},
		"parse subnet error": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "notAnIP"}}, nil),
			},
			wantErr: true,
		},
		"list servers error": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: stubPager{allPagesErr: someErr},
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"extract servers error": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{}, someErr),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"sever with empty name skipped": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						ID:        "id1",
						Tags:      &[]string{"constellation-role-control-plane", "constellation-uid-7777"},
						Addresses: newTestAddrs("192.0.2.5", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"sever with nil tags skipped": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						ID:        "id1",
						Addresses: newTestAddrs("192.0.2.5", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"server with empty ID skipped": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						Tags:      &[]string{"constellation-role-control-plane", "constellation-uid-7777"},
						Addresses: newTestAddrs("192.0.2.5", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"server with unknown role skipped": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						ID:        "id1",
						Tags:      &[]string{"constellation-role-unknown", "constellation-uid-7777"},
						Addresses: newTestAddrs("192.0.2.5", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"server without role skipped": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						ID:        "id1",
						Tags:      &[]string{"constellation-uid-7777"},
						Addresses: newTestAddrs("192.0.2.5", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"server without parseable addresses skipped": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						ID:        "id1",
						Tags:      &[]string{"constellation-role-control-plane", "constellation-uid-7777"},
						Addresses: map[string]any{"foo": "bar"},
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
		"server addresses contains in": {
			imds: &stubIMDSClient{uidResult: "uid"},
			api: &stubServersClient{
				serversPager: newSeverPager([]servers.Server{
					{
						Name:      "name1",
						ID:        "id1",
						Tags:      &[]string{"constellation-role-control-plane", "constellation-uid-7777"},
						Addresses: newTestAddrs("invalidIP", ""),
					},
				}, nil),
				netsPager:    newNetPager([]networks.Network{{Name: "mynet"}}, nil),
				subnetsPager: newSubnetPager([]subnets.Subnet{{Name: "mynet", CIDR: "192.0.2.0/24"}}, nil),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &MetadataClient{imds: tc.imds, api: tc.api}

			got, err := c.List(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

func TestUID(t *testing.T) {
	testCases := map[string]struct {
		imds    *stubIMDSClient
		want    string
		wantErr bool
	}{
		"error returned from IMDS client": {
			imds:    &stubIMDSClient{uidErr: errors.New("failed")},
			wantErr: true,
		},
		"UID returned from IMDS client": {
			imds: &stubIMDSClient{uidResult: "uid"},
			want: "uid",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &MetadataClient{imds: tc.imds}

			got, err := c.UID(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

func TestInitSecretHash(t *testing.T) {
	testCases := map[string]struct {
		imds    *stubIMDSClient
		want    []byte
		wantErr bool
	}{
		"error returned from IMDS client": {
			imds:    &stubIMDSClient{initSecretHashErr: errors.New("failed")},
			wantErr: true,
		},
		"initSecretHash returned from IMDS client": {
			imds: &stubIMDSClient{initSecretHashResult: "initSecretHash"},
			want: []byte("initSecretHash"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &MetadataClient{imds: tc.imds}

			got, err := c.InitSecretHash(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

func TestGetLoadBalancerEndpoint(t *testing.T) {
	testCases := map[string]struct {
		imds    *stubIMDSClient
		want    string
		wantErr bool
	}{
		"error returned from IMDS client": {
			imds:    &stubIMDSClient{loadBalancerEndpointErr: errors.New("failed")},
			wantErr: true,
		},
		"UID returned from IMDS client": {
			imds: &stubIMDSClient{loadBalancerEndpointResult: "some.endpoint"},
			want: "some.endpoint",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &MetadataClient{imds: tc.imds}

			got, _, err := c.GetLoadBalancerEndpoint(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}

// newNetPager returns a network pager as we would get from a ListNetworks.
func newNetPager(nets []networks.Network, err error) stubPager {
	return stubPager{
		page: networks.NetworkPage{
			LinkedPageBase: pagination.LinkedPageBase{
				PageResult: pagination.PageResult{
					Result: gophercloud.Result{
						Body: struct {
							Networks []networks.Network `json:"networks"`
						}{nets},
						Err: err,
					},
				},
			},
		},
	}
}

// newSubnetPager returns a subnet pager as we would get from a ListSubnets.
func newSubnetPager(nets []subnets.Subnet, err error) stubPager {
	return stubPager{
		page: subnets.SubnetPage{
			LinkedPageBase: pagination.LinkedPageBase{
				PageResult: pagination.PageResult{
					Result: gophercloud.Result{
						Body: struct {
							Subnets []subnets.Subnet `json:"subnets"`
						}{nets},
						Err: err,
					},
				},
			},
		},
	}
}

// newSeverPager returns a server pager as we would get from a ListServers.
func newSeverPager(srvs []servers.Server, err error) stubPager {
	return stubPager{
		page: servers.ServerPage{
			LinkedPageBase: pagination.LinkedPageBase{
				PageResult: pagination.PageResult{
					Result: gophercloud.Result{
						Body: struct {
							Servers []servers.Server `json:"servers"`
						}{srvs},
						Err: err,
					},
				},
			},
		},
	}
}
