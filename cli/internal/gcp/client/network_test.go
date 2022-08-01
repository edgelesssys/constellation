package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	"google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestCreateVPCs(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		operationGlobalAPI operationGlobalAPI
		operationRegionAPI operationRegionAPI
		networksAPI        networksAPI
		subnetworksAPI     subnetworksAPI
		wantErr            bool
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
			wantErr:            true,
		},
		"failed wait region op": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{waitErr: someErr},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			wantErr:            true,
		},
		"failed insert networks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{insertErr: someErr},
			subnetworksAPI:     stubSubnetworksAPI{},
			wantErr:            true,
		},
		"failed insert subnetworks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{insertErr: someErr},
			wantErr:            true,
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
				workers:            make(cloudtypes.Instances),
				controlPlanes:      make(cloudtypes.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateVPCs(ctx))
			} else {
				assert.NoError(client.CreateVPCs(ctx))
				assert.NotNil(client.network)
			}
		})
	}
}

func TestTerminateVPCs(t *testing.T) {
	someErr := errors.New("failed")
	notFoundErr := &googleapi.Error{Code: http.StatusNotFound}

	testCases := map[string]struct {
		operationGlobalAPI operationGlobalAPI
		operationRegionAPI operationRegionAPI
		networksAPI        networksAPI
		subnetworksAPI     subnetworksAPI
		firewalls          []string
		subnetwork         string
		network            string
		wantErr            bool
	}{
		"successful terminate": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
		},
		"subnetwork empty": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "",
			network:            "network-id-1",
		},
		"network empty": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "subnetwork-id-1",
			network:            "",
		},
		"subnetwork not found": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{deleteErr: notFoundErr},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
		},
		"network not found": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{deleteErr: notFoundErr},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
		},
		"failed wait global op": {
			operationGlobalAPI: stubOperationGlobalAPI{waitErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
			wantErr:            true,
		},
		"failed delete networks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{deleteErr: someErr},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
			wantErr:            true,
		},
		"failed delete subnetworks": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{deleteErr: someErr},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
			wantErr:            true,
		},
		"must delete firewalls first": {
			firewalls:          []string{"firewall-1", "firewall-2"},
			operationRegionAPI: stubOperationRegionAPI{},
			operationGlobalAPI: stubOperationGlobalAPI{},
			networksAPI:        stubNetworksAPI{},
			subnetworksAPI:     stubSubnetworksAPI{},
			subnetwork:         "subnetwork-id-1",
			network:            "network-id-1",
			wantErr:            true,
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
				network:            tc.network,
				subnetwork:         tc.subnetwork,
			}

			if tc.wantErr {
				assert.Error(client.TerminateVPCs(ctx))
			} else {
				assert.NoError(client.TerminateVPCs(ctx))
				assert.Empty(client.network)
				assert.Empty(client.subnetwork)
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
				FromPort:    9000,
			},
			cloudtypes.FirewallRule{
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
		network            string
		operationGlobalAPI operationGlobalAPI
		firewallsAPI       firewallsAPI
		firewallInput      FirewallInput
		wantErr            bool
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
			wantErr:            true,
		},
		"failed insert networks": {
			network:            "network",
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{insertErr: someErr},
			wantErr:            true,
		},
		"no network set": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{},
			wantErr:            true,
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

			if tc.wantErr {
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
	notFoundErr := &googleapi.Error{Code: http.StatusNotFound}

	testCases := map[string]struct {
		operationGlobalAPI operationGlobalAPI
		firewallsAPI       firewallsAPI
		firewalls          []string
		wantErr            bool
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
		"successful terminate when firewall not found": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{deleteErr: notFoundErr},
			firewalls:          []string{"firewall-1", "firewall-2"},
		},
		"failed to wait on global operation": {
			operationGlobalAPI: stubOperationGlobalAPI{waitErr: someErr},
			firewallsAPI:       stubFirewallsAPI{},
			firewalls:          []string{"firewall-1", "firewall-2"},
			wantErr:            true,
		},
		"failed to delete firewalls": {
			operationGlobalAPI: stubOperationGlobalAPI{},
			firewallsAPI:       stubFirewallsAPI{deleteErr: someErr},
			firewalls:          []string{"firewall-1", "firewall-2"},
			wantErr:            true,
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

			if tc.wantErr {
				assert.Error(client.TerminateFirewall(ctx))
			} else {
				assert.NoError(client.TerminateFirewall(ctx))
				assert.Empty(client.firewalls)
			}
		})
	}
}

func TestCreateLoadBalancer(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		operationRegionAPI operationRegionAPI
		healthChecksAPI    healthChecksAPI
		backendServicesAPI backendServicesAPI
		forwardingRulesAPI forwardingRulesAPI
		wantErr            bool
	}{
		"successful create": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{},
		},
		"CreateLoadBalancer fails when getting forwarding rule": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{getErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"CreateLoadBalancer fails when label fingerprint is missing": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{}},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"CreateLoadBalancer fails when creating health check": {
			healthChecksAPI:    stubHealthChecksAPI{insertErr: someErr},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"CreateLoadBalancer fails when creating backend service": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{insertErr: someErr},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"CreateLoadBalancer fails when creating forwarding rule": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{insertErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"CreateLoadBalancer fails when waiting on operation": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{waitErr: someErr},
			wantErr:            true,
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
				backendServicesAPI: tc.backendServicesAPI,
				forwardingRulesAPI: tc.forwardingRulesAPI,
				healthChecksAPI:    tc.healthChecksAPI,
				operationRegionAPI: tc.operationRegionAPI,
			}

			if tc.wantErr {
				assert.Error(client.CreateLoadBalancer(ctx))
			} else {
				assert.NoError(client.CreateLoadBalancer(ctx))
				assert.NotEmpty(client.healthCheck)
				assert.NotEmpty(client.backendService)
				assert.NotEmpty(client.forwardingRule)
			}
		})
	}
}

func TestTerminateLoadBalancer(t *testing.T) {
	someErr := errors.New("failed")
	notFoundErr := &googleapi.Error{Code: http.StatusNotFound}

	testCases := map[string]struct {
		operationRegionAPI operationRegionAPI
		healthChecksAPI    healthChecksAPI
		backendServicesAPI backendServicesAPI
		forwardingRulesAPI forwardingRulesAPI
		wantErr            bool
	}{
		"successful terminate": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
		},
		"successful terminate when health check not found": {
			healthChecksAPI:    stubHealthChecksAPI{deleteErr: notFoundErr},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
		},
		"successful terminate when backend service not found": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{deleteErr: notFoundErr},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
		},
		"successful terminate when forwarding rule not found": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{deleteErr: notFoundErr},
			operationRegionAPI: stubOperationRegionAPI{},
		},
		"TerminateLoadBalancer fails when deleting health check": {
			healthChecksAPI:    stubHealthChecksAPI{deleteErr: someErr},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"TerminateLoadBalancer fails when deleting backend service": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{deleteErr: someErr},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"TerminateLoadBalancer fails when deleting forwarding rule": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{deleteErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
		},
		"TerminateLoadBalancer fails when waiting on operation": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{waitErr: someErr},
			wantErr:            true,
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
				backendServicesAPI: tc.backendServicesAPI,
				forwardingRulesAPI: tc.forwardingRulesAPI,
				healthChecksAPI:    tc.healthChecksAPI,
				operationRegionAPI: tc.operationRegionAPI,
			}

			if tc.wantErr {
				assert.Error(client.TerminateLoadBalancer(ctx))
			} else {
				assert.NoError(client.TerminateLoadBalancer(ctx))
				assert.Empty(client.healthCheck)
				assert.Empty(client.backendService)
				assert.Empty(client.forwardingRule)
			}
		})
	}
}
