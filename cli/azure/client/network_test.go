package client

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
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
			networksAPI: stubNetworksAPI{stubResponse: stubVirtualNetworksCreateOrUpdatePollerResponse{pollerErr: someErr}},
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
				nodes:         make(azure.Instances),
				coordinators:  make(azure.Instances),
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
			networkSecurityGroupsAPI: stubNetworkSecurityGroupsAPI{stubPoller: stubNetworkSecurityGroupsCreateOrUpdatePollerResponse{pollerErr: someErr}},
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
				nodes:                    make(azure.Instances),
				coordinators:             make(azure.Instances),
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

// TODO: deprecate as soon as scale sets are available.
func TestCreateNIC(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		networkInterfacesAPI networkInterfacesAPI
		name                 string
		publicIPAddressID    string
		wantErr              bool
	}{
		"successful create": {
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			name:                 "nic-name",
			publicIPAddressID:    "pubIP-id",
		},
		"failed to get response from successful create": {
			networkInterfacesAPI: stubNetworkInterfacesAPI{stubResp: stubInterfacesClientCreateOrUpdatePollerResponse{pollErr: someErr}},
			wantErr:              true,
		},
		"failed create": {
			networkInterfacesAPI: stubNetworkInterfacesAPI{createErr: someErr},
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
				nodes:                make(azure.Instances),
				coordinators:         make(azure.Instances),
				networkInterfacesAPI: tc.networkInterfacesAPI,
			}

			ip, id, err := client.createNIC(ctx, tc.name, tc.publicIPAddressID)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotEmpty(ip)
				assert.NotEmpty(id)
			}
		})
	}
}

// TODO: deprecate as soon as scale sets are available.
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
			publicIPAddressesAPI: stubPublicIPAddressesAPI{stubCreateResponse: stubPublicIPAddressesClientCreateOrUpdatePollerResponse{pollErr: someErr}},
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
				nodes:                make(azure.Instances),
				coordinators:         make(azure.Instances),
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
