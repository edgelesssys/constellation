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

// ServiceAccountCreator creates service accounts.
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
		return "", state.ConstellationState{}, fmt.Errorf("creating service account not supported for GCP")
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

func (c *ServiceAccountCreator) createServiceAccountAzure(ctx context.Context, cl azureclient,
	stat state.ConstellationState, _ *config.Config,
) (string, state.ConstellationState, error) {
	cl.SetState(stat)

	serviceAccount, err := cl.CreateServicePrincipal(ctx)
	if err != nil {
		return "", state.ConstellationState{}, fmt.Errorf("creating service account: %w", err)
	}

	return serviceAccount, cl.GetState(), nil
}
