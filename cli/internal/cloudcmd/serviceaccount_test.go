package cloudcmd

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountCreator(t *testing.T) {
	someGCPState := func() state.ConstellationState {
		return state.ConstellationState{
			CloudProvider:                   cloudprovider.GCP.String(),
			GCPProject:                      "project",
			GCPWorkers:                      cloudtypes.Instances{},
			GCPControlPlanes:                cloudtypes.Instances{},
			GCPWorkerInstanceGroup:          "workers-group",
			GCPControlPlaneInstanceGroup:    "controlplane-group",
			GCPWorkerInstanceTemplate:       "template",
			GCPControlPlaneInstanceTemplate: "template",
			GCPNetwork:                      "network",
			GCPFirewalls:                    []string{},
		}
	}
	someAzureState := func() state.ConstellationState {
		return state.ConstellationState{
			CloudProvider: cloudprovider.Azure.String(),
		}
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		newGCPClient     func(ctx context.Context) (gcpclient, error)
		newAzureClient   func(subscriptionID, tenantID string) (azureclient, error)
		state            state.ConstellationState
		config           *config.Config
		wantErr          bool
		wantStateMutator func(*state.ConstellationState)
	}{
		"gcp": {
			newGCPClient: func(ctx context.Context) (gcpclient, error) {
				return &fakeGcpClient{}, nil
			},
			state:  someGCPState(),
			config: config.Default(),
			wantStateMutator: func(stat *state.ConstellationState) {
				stat.GCPServiceAccount = "service-account@project.iam.gserviceaccount.com"
			},
		},
		"gcp newGCPClient error": {
			newGCPClient: func(ctx context.Context) (gcpclient, error) {
				return nil, someErr
			},
			state:   someGCPState(),
			config:  config.Default(),
			wantErr: true,
		},
		"gcp client setState error": {
			newGCPClient: func(ctx context.Context) (gcpclient, error) {
				return &stubGcpClient{setStateErr: someErr}, nil
			},
			state:   someGCPState(),
			config:  config.Default(),
			wantErr: true,
		},
		"gcp client createServiceAccount error": {
			newGCPClient: func(ctx context.Context) (gcpclient, error) {
				return &stubGcpClient{createServiceAccountErr: someErr}, nil
			},
			state:   someGCPState(),
			config:  config.Default(),
			wantErr: true,
		},
		"gcp client getState error": {
			newGCPClient: func(ctx context.Context) (gcpclient, error) {
				return &stubGcpClient{getStateErr: someErr}, nil
			},
			state:   someGCPState(),
			config:  config.Default(),
			wantErr: true,
		},
		"azure": {
			newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
				return &fakeAzureClient{}, nil
			},
			state:  someAzureState(),
			config: config.Default(),
			wantStateMutator: func(stat *state.ConstellationState) {
				stat.AzureADAppObjectID = "00000000-0000-0000-0000-000000000001"
			},
		},
		"azure newAzureClient error": {
			newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
				return nil, someErr
			},
			state:   someAzureState(),
			config:  config.Default(),
			wantErr: true,
		},
		"azure client setState error": {
			newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
				return &stubAzureClient{setStateErr: someErr}, nil
			},
			state:   someAzureState(),
			config:  config.Default(),
			wantErr: true,
		},
		"azure client createServiceAccount error": {
			newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
				return &stubAzureClient{createServicePrincipalErr: someErr}, nil
			},
			state:   someAzureState(),
			config:  config.Default(),
			wantErr: true,
		},
		"azure client getState error": {
			newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
				return &stubAzureClient{getStateErr: someErr}, nil
			},
			state:   someAzureState(),
			config:  config.Default(),
			wantErr: true,
		},
		"qemu": {
			state:            state.ConstellationState{CloudProvider: "qemu"},
			wantStateMutator: func(cs *state.ConstellationState) {},
			config:           config.Default(),
		},
		"unknown cloud provider": {
			state:   state.ConstellationState{},
			config:  config.Default(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			creator := &ServiceAccountCreator{
				newGCPClient:   tc.newGCPClient,
				newAzureClient: tc.newAzureClient,
			}

			serviceAccount, state, err := creator.Create(context.Background(), tc.state, tc.config)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotEmpty(serviceAccount)
				tc.wantStateMutator(&tc.state)
				assert.Equal(tc.state, state)
			}
		})
	}
}
