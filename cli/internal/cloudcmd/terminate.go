package cloudcmd

import (
	"context"
	"fmt"

	azurecl "github.com/edgelesssys/constellation/cli/internal/azure/client"
	gcpcl "github.com/edgelesssys/constellation/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/state"
)

// Terminator deletes cloud provider resources.
type Terminator struct {
	newGCPClient   func(ctx context.Context) (gcpclient, error)
	newAzureClient func(subscriptionID, tenantID string) (azureclient, error)
}

// NewTerminator create a new cloud terminator.
func NewTerminator() *Terminator {
	return &Terminator{
		newGCPClient: func(ctx context.Context) (gcpclient, error) {
			return gcpcl.NewFromDefault(ctx)
		},
		newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
			return azurecl.NewFromDefault(subscriptionID, tenantID)
		},
	}
}

// Terminate deletes the could provider resources defined in the constellation state.
func (t *Terminator) Terminate(ctx context.Context, state state.ConstellationState) error {
	provider := cloudprovider.FromString(state.CloudProvider)
	switch provider {
	case cloudprovider.GCP:
		cl, err := t.newGCPClient(ctx)
		if err != nil {
			return err
		}
		defer cl.Close()
		return t.terminateGCP(ctx, cl, state)
	case cloudprovider.Azure:
		cl, err := t.newAzureClient(state.AzureSubscription, state.AzureTenant)
		if err != nil {
			return err
		}
		return t.terminateAzure(ctx, cl, state)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (t *Terminator) terminateGCP(ctx context.Context, cl gcpclient, state state.ConstellationState) error {
	cl.SetState(state)

	if err := cl.TerminateLoadBalancers(ctx); err != nil {
		return err
	}
	if err := cl.TerminateInstances(ctx); err != nil {
		return err
	}
	if err := cl.TerminateFirewall(ctx); err != nil {
		return err
	}
	if err := cl.TerminateVPCs(ctx); err != nil {
		return err
	}

	return nil
}

func (t *Terminator) terminateAzure(ctx context.Context, cl azureclient, state state.ConstellationState) error {
	cl.SetState(state)

	return cl.TerminateResourceGroupResources(ctx)
}
