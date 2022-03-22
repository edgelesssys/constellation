package client

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/stretchr/testify/assert"
)

func TestCreateVPCs(t *testing.T) {
	someErr := errors.New("failed")

	testInput := VPCsInput{
		SubnetCIDR:    "192.0.2.0/24",
		SubnetExtCIDR: "198.51.100.0/24",
	}

	testCases := map[string]struct {
		operationGlobalAPI operationGlobalAPI
		operationRegionAPI operationRegionAPI
		networksAPI        networksAPI
		subnetworksAPI     subnetworksAPI
		errExpected        bool
	}{
		"successful create": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
		},
		"failed wait global op": {
			operationGlobalAPI: stubOperationGlobalAPI{waitErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			errExpected:        true,
		},
		"failed wait region op": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{waitErr: someErr},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			errExpected:        true,
		},
		"failed insert networks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{insertErr: someErr},
			subnetworksAPI:     stubSubnetworksAPI{},
			errExpected:        true,
		},
		"failed insert subnetworks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{insertErr: someErr},
			errExpected:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				operationGlobalAPI: tc.operationGlobalAPI,
				operationRegionAPI: tc.operationRegionAPI,
				networksAPI:        tc.networksAPI,
				subnetworksAPI:     tc.subnetworksAPI,
				nodes:              make(gcp.Instances),
				coordinators:       make(gcp.Instances),
			}

			if tc.errExpected {
				assert.Error(client.CreateVPCs(ctx, testInput))
			} else {
				assert.NoError(client.CreateVPCs(ctx, testInput))
				assert.NotNil(client.network)
			}
		})
	}
}

func TestTerminateVPCs(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		operationGlobalAPI operationGlobalAPI
		operationRegionAPI operationRegionAPI
		networksAPI        networksAPI
		subnetworksAPI     subnetworksAPI
		firewalls          []string
		errExpected        bool
	}{
		"successful terminate": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
		},
		"failed wait global op": {
			operationGlobalAPI: stubOperationGlobalAPI{waitErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			errExpected:        true,
		},
		"failed delete networks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{deleteErr: someErr},
			subnetworksAPI:     stubSubnetworksAPI{},
			errExpected:        true,
		},
		"failed delete subnetworks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{deleteErr: someErr},
			errExpected:        true,
		},
		"must delete firewalls first": {
			firewalls:          []string{"firewall-1", "firewall-2"},
			operationRegionAPI: stubOperationRegionAPI{},
			operationGlobalAPI: stubOperationGlobalAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			errExpected:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				operationGlobalAPI: tc.operationGlobalAPI,
				operationRegionAPI: tc.operationRegionAPI,
				networksAPI:        tc.networksAPI,
				subnetworksAPI:     tc.subnetworksAPI,
				firewalls:          tc.firewalls,
				network:            "network-id-1",
				subnetwork:         "subnetwork-id-1",
			}

			if tc.errExpected {
				assert.Error(client.TerminateVPCs(ctx))
			} else {
				assert.NoError(client.TerminateVPCs(ctx))
				assert.Empty(client.network)
			}
		})
	}
}

func TestCreateFirewall(t *testing.T) {
	someErr := errors.New("failed")
	testFirewallInput := FirewallInput{
		Ingress: cloudtypes.Firewall{
			cloudtypes.FirewallRule{
				Name:        "test-1",
				Description: "test-1 description",
				Protocol:    "tcp",
				IPRange:     "192.0.2.0/24",
				Port:        9000,
			},
			cloudtypes.FirewallRule{
				Name:        "test-2",
				Description: "test-2 description",
				Protocol:    "udp",
				IPRange:     "192.0.2.0/24",
				Port:        51820,
			},
		},
		Egress: cloudtypes.Firewall{},
	}

	testCases := map[string]struct {
		network            string
		operationGlobalAPI operationGlobalAPI
		firewallsAPI       firewallsAPI
		firewallInput      FirewallInput
		errExpected        bool
	}{
		"successful create": {
			network:            "network",
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{},
		},
		"failed wait global op": {
			network:            "network",
			operationGlobalAPI: stubOperationGlobalAPI{waitErr: someErr},
			firewallsAPI:       stubFirewallsAPI{},
			errExpected:        true,
		},
		"failed insert networks": {
			network:            "network",
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{insertErr: someErr},
			errExpected:        true,
		},
		"no network set": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{},
			errExpected:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				network:            tc.network,
				operationGlobalAPI: tc.operationGlobalAPI,
				firewallsAPI:       tc.firewallsAPI,
			}

			if tc.errExpected {
				assert.Error(client.CreateFirewall(ctx, testFirewallInput))
			} else {
				assert.NoError(client.CreateFirewall(ctx, testFirewallInput))
				assert.ElementsMatch([]string{"test-1", "test-2"}, client.firewalls)
			}
		})
	}
}

func TestTerminateFirewall(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		operationGlobalAPI operationGlobalAPI
		firewallsAPI       firewallsAPI
		firewalls          []string
		errExpected        bool
	}{
		"successful terminate": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{},
			firewalls:          []string{"firewall-1", "firewall-2"},
		},
		"successful terminate when no firewall exists": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{},
			firewalls:          []string{},
		},
		"failed to wait on global operation": {
			operationGlobalAPI: stubOperationGlobalAPI{waitErr: someErr},
			firewallsAPI:       stubFirewallsAPI{},
			firewalls:          []string{"firewall-1", "firewall-2"},
			errExpected:        true,
		},
		"failed to delete firewalls": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{deleteErr: someErr},
			firewalls:          []string{"firewall-1", "firewall-2"},
			errExpected:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				firewalls:          tc.firewalls,
				operationGlobalAPI: tc.operationGlobalAPI,
				firewallsAPI:       tc.firewallsAPI,
			}

			if tc.errExpected {
				assert.Error(client.TerminateFirewall(ctx))
			} else {
				assert.NoError(client.TerminateFirewall(ctx))
				assert.Empty(client.firewalls)
			}
		})
	}
}
