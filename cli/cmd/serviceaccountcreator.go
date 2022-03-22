package cmd

import (
	"context"
	"fmt"

	azurecl "github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	ec2cl "github.com/edgelesssys/constellation/cli/ec2/client"
	gcpcl "github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

type serviceAccountCreator interface {
	createServiceAccount(ctx context.Context, stat state.ConstellationState, config *config.Config) (string, state.ConstellationState, error)
}

type serviceAccountClient struct{}

// createServiceAccount creates a new cloud provider service account with access to the created resources.
func (c serviceAccountClient) createServiceAccount(ctx context.Context, stat state.ConstellationState, config *config.Config) (string, state.ConstellationState, error) {
	switch stat.CloudProvider {
	case cloudprovider.AWS.String():
		// TODO: implement
		ec2client, err := ec2cl.NewFromDefault(ctx)
		if err != nil {
			return "", state.ConstellationState{}, err
		}
		return c.createServiceAccountEC2(ctx, ec2client, stat, config)
	case cloudprovider.GCP.String():
		gcpclient, err := gcpcl.NewFromDefault(ctx)
		if err != nil {
			return "", state.ConstellationState{}, err
		}
		serviceAccount, stat, err := c.createServiceAccountGCP(ctx, gcpclient, stat, config)
		if err != nil {
			return "", state.ConstellationState{}, err
		}
		return serviceAccount, stat, gcpclient.Close()
	case cloudprovider.Azure.String():
		azureclient, err := azurecl.NewFromDefault(stat.AzureSubscription, stat.AzureTenant)
		if err != nil {
			return "", state.ConstellationState{}, err
		}
		return c.createServiceAccountAzure(ctx, azureclient, stat)
	}

	return "", state.ConstellationState{}, fmt.Errorf("unknown cloud provider %v", stat.CloudProvider)
}

func (c serviceAccountClient) createServiceAccountAzure(ctx context.Context, cl azureclient, stat state.ConstellationState) (string, state.ConstellationState, error) {
	if err := cl.SetState(stat); err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to set state while creating service account: %w", err)
	}
	serviceAccount, err := cl.CreateServicePrincipal(ctx)
	if err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to create service account: %w", err)
	}

	stat, err = cl.GetState()
	if err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to get state after creating service account: %w", err)
	}
	return serviceAccount, stat, nil
}

func (c serviceAccountClient) createServiceAccountGCP(ctx context.Context, cl gcpclient, stat state.ConstellationState, config *config.Config) (string, state.ConstellationState, error) {
	if err := cl.SetState(stat); err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to set state while creating service account: %w", err)
	}

	input := gcpcl.ServiceAccountInput{
		Roles: *config.Provider.GCP.ServiceAccountRoles,
	}
	serviceAccount, err := cl.CreateServiceAccount(ctx, input)
	if err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to create service account: %w", err)
	}

	stat, err = cl.GetState()
	if err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to get state after creating service account: %w", err)
	}
	return serviceAccount, stat, nil
}

//nolint:unparam
func (c serviceAccountClient) createServiceAccountEC2(ctx context.Context, cl ec2client, stat state.ConstellationState, config *config.Config) (string, state.ConstellationState, error) {
	// TODO: implement
	if err := cl.SetState(stat); err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to set state while creating service account: %w", err)
	}
	return "", stat, nil
}
