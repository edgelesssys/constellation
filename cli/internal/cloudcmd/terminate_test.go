package cloudcmd

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	someGCPState := func() state.ConstellationState {
		return state.ConstellationState{
			CloudProvider: cloudprovider.GCP.String(),
			GCPProject:    "project",
			GCPWorkerInstances: cloudtypes.Instances{
				"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
				"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			},
			GCPControlPlaneInstances: cloudtypes.Instances{
				"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			},
			GCPWorkerInstanceGroup:          "worker-group",
			GCPControlPlaneInstanceGroup:    "controlplane-group",
			GCPWorkerInstanceTemplate:       "template",
			GCPControlPlaneInstanceTemplate: "template",
			GCPNetwork:                      "network",
			GCPFirewalls:                    []string{"a", "b", "c"},
		}
	}
	someAzureState := func() state.ConstellationState {
		return state.ConstellationState{
			CloudProvider: cloudprovider.Azure.String(),
			AzureWorkerInstances: cloudtypes.Instances{
				"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
				"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			},
			AzureControlPlaneInstances: cloudtypes.Instances{
				"id-c": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			},
			AzureADAppObjectID: "00000000-0000-0000-0000-000000000001",
		}
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		gcpclient         gcpclient
		newGCPClientErr   error
		azureclient       azureclient
		newAzureClientErr error
		state             state.ConstellationState
		wantErr           bool
	}{
		"gcp": {
			gcpclient: &stubGcpClient{},
			state:     someGCPState(),
		},
		"gcp newGCPClient error": {
			newGCPClientErr: someErr,
			state:           someGCPState(),
			wantErr:         true,
		},
		"gcp terminateInstances error": {
			gcpclient: &stubGcpClient{terminateInstancesErr: someErr},
			state:     someGCPState(),
			wantErr:   true,
		},
		"gcp terminateFirewall error": {
			gcpclient: &stubGcpClient{terminateFirewallErr: someErr},
			state:     someGCPState(),
			wantErr:   true,
		},
		"gcp terminateVPCs error": {
			gcpclient: &stubGcpClient{terminateVPCsErr: someErr},
			state:     someGCPState(),
			wantErr:   true,
		},
		"azure": {
			azureclient: &stubAzureClient{},
			state:       someAzureState(),
		},
		"azure newAzureClient error": {
			newAzureClientErr: someErr,
			state:             someAzureState(),
			wantErr:           true,
		},
		"azure terminateResourceGroupResources error": {
			azureclient: &stubAzureClient{terminateResourceGroupResourcesErr: someErr},
			state:       someAzureState(),
			wantErr:     true,
		},
		"unknown cloud provider": {
			state:   state.ConstellationState{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			terminator := &Terminator{
				newGCPClient: func(ctx context.Context) (gcpclient, error) {
					return tc.gcpclient, tc.newGCPClientErr
				},
				newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
					return tc.azureclient, tc.newAzureClientErr
				},
			}

			err := terminator.Terminate(context.Background(), tc.state)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				switch cloudprovider.FromString(tc.state.CloudProvider) {
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
		})
	}
}
