/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestCreator(t *testing.T) {
	wantGCPState := state.ConstellationState{
		CloudProvider: cloudprovider.GCP.String(),
		GCPProject:    "project",
		GCPControlPlaneInstances: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		GCPWorkerInstances: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-2": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		GCPWorkerInstanceGroup:          "workers-group",
		GCPControlPlaneInstanceGroup:    "controlplane-group",
		GCPWorkerInstanceTemplate:       "worker-template",
		GCPControlPlaneInstanceTemplate: "controlplane-template",
		GCPNetwork:                      "network",
		GCPSubnetwork:                   "subnetwork",
		GCPLoadbalancers:                []string{"kube-lb", "boot-lb", "verify-lb"},
		GCPFirewalls: []string{
			"bootstrapper", "ssh", "nodeport", "kubernetes", "konnectivity",
			"allow-cluster-internal-tcp", "allow-cluster-internal-udp", "allow-cluster-internal-icmp",
			"allow-node-internal-tcp", "allow-node-internal-udp", "allow-node-internal-icmp",
		},
	}

	wantAzureState := state.ConstellationState{
		CloudProvider: cloudprovider.Azure.String(),
		AzureControlPlaneInstances: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureWorkerInstances: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-2": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureSubnet:               "subnet",
		AzureNetworkSecurityGroup: "network-security-group",
		AzureWorkerScaleSet:       "workers-scale-set",
		AzureControlPlaneScaleSet: "controlplanes-scale-set",
	}

	someErr := errors.New("failed")

	testCases := map[string]struct {
		gcpclient         gcpclient
		newGCPClientErr   error
		azureclient       azureclient
		newAzureClientErr error
		provider          cloudprovider.Provider
		config            *config.Config
		wantState         state.ConstellationState
		wantErr           bool
		wantRollback      bool // Use only together with stubClients.
	}{
		"gcp": {
			gcpclient: &fakeGcpClient{project: "project"},
			provider:  cloudprovider.GCP,
			config:    config.Default(),
			wantState: wantGCPState,
		},
		"gcp newGCPClient error": {
			newGCPClientErr: someErr,
			provider:        cloudprovider.GCP,
			config:          config.Default(),
			wantErr:         true,
		},
		"gcp CreateVPCs error": {
			gcpclient:    &stubGcpClient{createVPCsErr: someErr},
			provider:     cloudprovider.GCP,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"gcp CreateFirewall error": {
			gcpclient:    &stubGcpClient{createFirewallErr: someErr},
			provider:     cloudprovider.GCP,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"gcp CreateInstances error": {
			gcpclient:    &stubGcpClient{createInstancesErr: someErr},
			provider:     cloudprovider.GCP,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"gcp CreateLoadBalancer error": {
			gcpclient:    &stubGcpClient{createLoadBalancerErr: someErr},
			provider:     cloudprovider.GCP,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"azure": {
			azureclient: &fakeAzureClient{},
			provider:    cloudprovider.Azure,
			config:      config.Default(),
			wantState:   wantAzureState,
		},
		"azure newAzureClient error": {
			newAzureClientErr: someErr,
			provider:          cloudprovider.Azure,
			config:            config.Default(),
			wantErr:           true,
		},
		"azure CreateVirtualNetwork error": {
			azureclient:  &stubAzureClient{createVirtualNetworkErr: someErr},
			provider:     cloudprovider.Azure,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"azure CreateSecurityGroup error": {
			azureclient:  &stubAzureClient{createSecurityGroupErr: someErr},
			provider:     cloudprovider.Azure,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"azure CreateInstances error": {
			azureclient:  &stubAzureClient{createInstancesErr: someErr},
			provider:     cloudprovider.Azure,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"unknown provider": {
			provider: cloudprovider.Unknown,
			config:   config.Default(),
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			creator := &Creator{
				out: &bytes.Buffer{},
				newGCPClient: func(ctx context.Context, project, zone, region, name string) (gcpclient, error) {
					return tc.gcpclient, tc.newGCPClientErr
				},
				newAzureClient: func(subscriptionID, tenantID, name, location, resourceGroup string) (azureclient, error) {
					return tc.azureclient, tc.newAzureClientErr
				},
			}

			state, err := creator.Create(context.Background(), tc.provider, tc.config, "name", "type", 2, 3)

			if tc.wantErr {
				assert.Error(err)
				if tc.wantRollback {
					switch tc.provider {
					case cloudprovider.GCP:
						cl := tc.gcpclient.(*stubGcpClient)
						assert.True(cl.terminateFirewallCalled)
						assert.True(cl.terminateInstancesCalled)
						assert.True(cl.terminateVPCsCalled)
						assert.True(cl.closeCalled)
					case cloudprovider.Azure:
						cl := tc.azureclient.(*stubAzureClient)
						assert.True(cl.terminateResourceGroupResourcesCalled)
					}
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantState, state)
			}
		})
	}
}
