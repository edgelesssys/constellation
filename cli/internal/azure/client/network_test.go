/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/stretchr/testify/assert"
)

func TestCreateVirtualNetwork(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		networksAPI networksAPI
		wantErr     bool
	}{
		"successful create": {
			networksAPI: stubNetworksAPI{},
		},
		"failed to get response from successful create": {
			networksAPI: stubNetworksAPI{pollErr: someErr},
			wantErr:     true,
		},
		"failed create": {
			networksAPI: stubNetworksAPI{createErr: someErr},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				resourceGroup: "resource-group",
				location:      "location",
				name:          "name",
				uid:           "uid",
				networksAPI:   tc.networksAPI,
				workers:       make(cloudtypes.Instances),
				controlPlanes: make(cloudtypes.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateVirtualNetwork(ctx))
			} else {
				assert.NoError(client.CreateVirtualNetwork(ctx))
				assert.NotEmpty(client.subnetID)
			}
		})
	}
}

func TestCreateSecurityGroup(t *testing.T) {
	someErr := errors.New("failed")
	testNetworkSecurityGroupInput := NetworkSecurityGroupInput{
		Ingress: cloudtypes.Firewall{
			{
				Name:        "test-1",
				Description: "test-1 description",
				Protocol:    "tcp",
				IPRange:     "192.0.2.0/24",
				FromPort:    9000,
			},
			{
				Name:        "test-2",
				Description: "test-2 description",
				Protocol:    "udp",
				IPRange:     "192.0.2.0/24",
				FromPort:    51820,
			},
		},
		Egress: cloudtypes.Firewall{},
	}

	testCases := map[string]struct {
		networkSecurityGroupsAPI networkSecurityGroupsAPI
		wantErr                  bool
	}{
		"successful create": {
			networkSecurityGroupsAPI: stubNetworkSecurityGroupsAPI{},
		},
		"failed to get response from successful create": {
			networkSecurityGroupsAPI: stubNetworkSecurityGroupsAPI{pollErr: someErr},
			wantErr:                  true,
		},
		"failed create": {
			networkSecurityGroupsAPI: stubNetworkSecurityGroupsAPI{createErr: someErr},
			wantErr:                  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				resourceGroup:            "resource-group",
				location:                 "location",
				name:                     "name",
				uid:                      "uid",
				workers:                  make(cloudtypes.Instances),
				controlPlanes:            make(cloudtypes.Instances),
				networkSecurityGroupsAPI: tc.networkSecurityGroupsAPI,
			}

			if tc.wantErr {
				assert.Error(client.CreateSecurityGroup(ctx, testNetworkSecurityGroupInput))
			} else {
				assert.NoError(client.CreateSecurityGroup(ctx, testNetworkSecurityGroupInput))
				assert.Equal("network-security-group-id", client.networkSecurityGroup)
			}
		})
	}
}

func TestCreatePublicIPAddress(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		publicIPAddressesAPI publicIPAddressesAPI
		name                 string
		wantErr              bool
	}{
		"successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			name:                 "nic-name",
		},
		"failed to get response from successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{pollErr: someErr},
			wantErr:              true,
		},
		"failed create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{createErr: someErr},
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				resourceGroup:        "resource-group",
				location:             "location",
				name:                 "name",
				uid:                  "uid",
				workers:              make(cloudtypes.Instances),
				controlPlanes:        make(cloudtypes.Instances),
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
			}

			id, err := client.createPublicIPAddress(ctx, tc.name)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotEmpty(id)
			}
		})
	}
}

func TestCreateExternalLoadBalancer(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		publicIPAddressesAPI publicIPAddressesAPI
		loadBalancersAPI     loadBalancersAPI
		isDebugCluster       bool
		wantErr              bool
	}{
		"successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			loadBalancersAPI:     stubLoadBalancersAPI{},
		},
		"successful create (debug cluster)": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			loadBalancersAPI:     stubLoadBalancersAPI{},
			isDebugCluster:       true,
		},
		"failed to get response from successful create": {
			loadBalancersAPI:     stubLoadBalancersAPI{pollErr: someErr},
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			wantErr:              true,
		},
		"failed create": {
			loadBalancersAPI:     stubLoadBalancersAPI{createErr: someErr},
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			wantErr:              true,
		},
		"cannot create public IP": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{createErr: someErr},
			loadBalancersAPI:     stubLoadBalancersAPI{},
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				resourceGroup:        "resource-group",
				location:             "location",
				name:                 "name",
				uid:                  "uid",
				workers:              make(cloudtypes.Instances),
				controlPlanes:        make(cloudtypes.Instances),
				loadBalancersAPI:     tc.loadBalancersAPI,
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
			}

			err := client.CreateExternalLoadBalancer(ctx, tc.isDebugCluster)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
