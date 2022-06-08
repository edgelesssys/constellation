package cloudcmd

import (
	"context"
	"fmt"

	azurecl "github.com/edgelesssys/constellation/cli/internal/azure/client"
	gcpcl "github.com/edgelesssys/constellation/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

// ServieAccountCreator creates service accounts.
type ServiceAccountCreator struct {
	newGCPClient   func(ctx context.Context) (gcpclient, error)
	newAzureClient func(subscriptionID, tenantID string) (azureclient, error)
}

func NewServiceAccountCreator() *ServiceAccountCreator {
	return &ServiceAccountCreator{
		newGCPClient: func(ctx context.Context) (gcpclient, error) {
			return gcpcl.NewFromDefault(ctx)
		},
		newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
			return azurecl.NewFromDefault(subscriptionID, tenantID)
		},
	}
}

// Create creates a new cloud provider service account with access to the created resources.
func (c *ServiceAccountCreator) Create(ctx context.Context, stat state.ConstellationState, config *config.Config,
) (string, state.ConstellationState, error) {
	provider := cloudprovider.FromString(stat.CloudProvider)
	switch provider {
	case cloudprovider.GCP:
		cl, err := c.newGCPClient(ctx)
		if err != nil {
			return "", state.ConstellationState{}, err
		}
		defer cl.Close()

		serviceAccount, stat, err := c.createServiceAccountGCP(ctx, cl, stat, config)
		if err != nil {
			return "", state.ConstellationState{}, err
		}

		return serviceAccount, stat, err
	case cloudprovider.Azure:
		cl, err := c.newAzureClient(stat.AzureSubscription, stat.AzureTenant)
		if err != nil {
			return "", state.ConstellationState{}, err
		}

		serviceAccount, stat, err := c.createServiceAccountAzure(ctx, cl, stat, config)
		if err != nil {
			return "", state.ConstellationState{}, err
		}

		return serviceAccount, stat, err
	case cloudprovider.QEMU:
		return "unsupported://qemu", stat, nil
	default:
		return "", state.ConstellationState{}, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (c *ServiceAccountCreator) createServiceAccountGCP(ctx context.Context, cl gcpclient,
	stat state.ConstellationState, config *config.Config,
) (string, state.ConstellationState, error) {
	if err := cl.SetState(stat); err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("failed to set state while creating service account: %w", err)
	}

	input := gcpcl.ServiceAccountInput{
		Roles: config.Provider.GCP.ServiceAccountRoles,
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

func (c *ServiceAccountCreator) createServiceAccountAzure(ctx context.Context, cl azureclient,
	stat state.ConstellationState, config *config.Config,
) (string, state.ConstellationState, error) {
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
