package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestCreateServiceAccountAzure(t *testing.T) {
	testState := state.ConstellationState{
		CloudProvider: cloudprovider.Azure.String(),
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState state.ConstellationState
		client        azureclient
		errExpected   bool
	}{
		"create service account works": {
			existingState: testState,
			client:        &fakeAzureClient{},
		},
		"fail setState": {
			existingState: testState,
			client:        &stubAzureClient{setStateErr: someErr},
			errExpected:   true,
		},
		"fail create": {
			existingState: testState,
			client:        &stubAzureClient{createServicePrincipalErr: someErr},
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			client := serviceAccountClient{}
			serviceAccount, _, err := client.createServiceAccountAzure(context.Background(), tc.client, tc.existingState)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotNil(serviceAccount)
				stat, err := tc.client.GetState()
				assert.NoError(err)
				assert.Equal(state.ConstellationState{
					CloudProvider:      cloudprovider.Azure.String(),
					AzureADAppObjectID: "00000000-0000-0000-0000-000000000001",
				}, stat)
			}
		})
	}
}

func TestCreateServiceAccountGCP(t *testing.T) {
	testState := state.ConstellationState{
		GCPProject:                     "project",
		GCPNodes:                       gcp.Instances{},
		GCPCoordinators:                gcp.Instances{},
		GCPNodeInstanceGroup:           "nodes-group",
		GCPCoordinatorInstanceGroup:    "coordinator-group",
		GCPNodeInstanceTemplate:        "template",
		GCPCoordinatorInstanceTemplate: "template",
		GCPNetwork:                     "network",
		GCPFirewalls:                   []string{},
	}
	config := config.Default()
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState state.ConstellationState
		client        gcpclient
		errExpected   bool
	}{
		"create service account works": {
			existingState: testState,
			client:        &fakeGcpClient{},
		},
		"fail setState": {
			existingState: testState,
			client:        &stubGcpClient{setStateErr: someErr},
			errExpected:   true,
		},
		"fail create": {
			existingState: testState,
			client:        &stubGcpClient{createServiceAccountErr: someErr},
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			client := serviceAccountClient{}
			serviceAccount, _, err := client.createServiceAccountGCP(context.Background(), tc.client, tc.existingState, config)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotNil(serviceAccount)
				stat, err := tc.client.GetState()
				assert.NoError(err)
				assert.Equal(state.ConstellationState{
					CloudProvider:                  cloudprovider.GCP.String(),
					GCPProject:                     "project",
					GCPNodes:                       gcp.Instances{},
					GCPCoordinators:                gcp.Instances{},
					GCPNodeInstanceGroup:           "nodes-group",
					GCPCoordinatorInstanceGroup:    "coordinator-group",
					GCPNodeInstanceTemplate:        "template",
					GCPCoordinatorInstanceTemplate: "template",
					GCPNetwork:                     "network",
					GCPFirewalls:                   []string{},
					GCPServiceAccount:              "service-account@project.iam.gserviceaccount.com",
				}, stat)
			}
		})
	}
}

type stubServiceAccountCreator struct {
	cloudServiceAccountURI string
	createErr              error
}

func (c *stubServiceAccountCreator) createServiceAccount(ctx context.Context, stat state.ConstellationState, config *config.Config) (string, state.ConstellationState, error) {
	return c.cloudServiceAccountURI, stat, c.createErr
}
