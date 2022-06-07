package cloudcmd

import (
	"context"

	azurecl "github.com/edgelesssys/constellation/cli/azure/client"
	gcpcl "github.com/edgelesssys/constellation/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type gcpclient interface {
	GetState() (state.ConstellationState, error)
	SetState(state.ConstellationState) error
	CreateVPCs(ctx context.Context) error
	CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error
	CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error
	CreateServiceAccount(ctx context.Context, input gcpcl.ServiceAccountInput) (string, error)
	TerminateFirewall(ctx context.Context) error
	TerminateVPCs(context.Context) error
	TerminateInstances(context.Context) error
	TerminateServiceAccount(ctx context.Context) error
	Close() error
}

type azureclient interface {
	GetState() (state.ConstellationState, error)
	SetState(state.ConstellationState) error
	CreateResourceGroup(ctx context.Context) error
	CreateExternalLoadBalancer(ctx context.Context) error
	CreateVirtualNetwork(ctx context.Context) error
	CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error
	CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error
	// TODO: deprecate as soon as scale sets are available
	CreateInstancesVMs(ctx context.Context, input azurecl.CreateInstancesInput) error
	CreateServicePrincipal(ctx context.Context) (string, error)
	TerminateResourceGroup(ctx context.Context) error
	TerminateServicePrincipal(ctx context.Context) error
}
