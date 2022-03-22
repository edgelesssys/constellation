package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type azureclient interface {
	GetState() (state.ConstellationState, error)
	SetState(state.ConstellationState) error
	CreateResourceGroup(ctx context.Context) error
	CreateVirtualNetwork(ctx context.Context) error
	CreateSecurityGroup(ctx context.Context, input client.NetworkSecurityGroupInput) error
	CreateInstances(ctx context.Context, input client.CreateInstancesInput) error
	// TODO: deprecate as soon as scale sets are available
	CreateInstancesVMs(ctx context.Context, input client.CreateInstancesInput) error
	CreateServicePrincipal(ctx context.Context) (string, error)
	TerminateResourceGroup(ctx context.Context) error
	TerminateServicePrincipal(ctx context.Context) error
}
