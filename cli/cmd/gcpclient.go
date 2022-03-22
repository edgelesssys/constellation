package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type gcpclient interface {
	GetState() (state.ConstellationState, error)
	SetState(state.ConstellationState) error
	CreateVPCs(ctx context.Context, input client.VPCsInput) error
	CreateFirewall(ctx context.Context, input client.FirewallInput) error
	CreateInstances(ctx context.Context, input client.CreateInstancesInput) error
	CreateServiceAccount(ctx context.Context, input client.ServiceAccountInput) (string, error)
	TerminateFirewall(ctx context.Context) error
	TerminateVPCs(context.Context) error
	TerminateInstances(context.Context) error
	TerminateServiceAccount(ctx context.Context) error
	Close() error
}
