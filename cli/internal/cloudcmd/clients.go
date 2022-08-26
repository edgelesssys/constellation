package cloudcmd

import (
	"context"

	azurecl "github.com/edgelesssys/constellation/cli/internal/azure/client"
	gcpcl "github.com/edgelesssys/constellation/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type gcpclient interface {
	GetState() state.ConstellationState
	SetState(state.ConstellationState)
	CreateVPCs(ctx context.Context) error
	CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error
	CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error
	CreateLoadBalancers(ctx context.Context) error
	TerminateFirewall(ctx context.Context) error
	TerminateVPCs(context.Context) error
	TerminateLoadBalancers(context.Context) error
	TerminateInstances(context.Context) error
	Close() error
}

type azureclient interface {
	GetState() state.ConstellationState
	SetState(state.ConstellationState)
	CreateApplicationInsight(ctx context.Context) error
	CreateResourceGroup(ctx context.Context) error
	CreateExternalLoadBalancer(ctx context.Context) error
	CreateVirtualNetwork(ctx context.Context) error
	CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error
	CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error
	CreateServicePrincipal(ctx context.Context) (string, error)
	TerminateResourceGroup(ctx context.Context) error
	TerminateServicePrincipal(ctx context.Context) error
}
