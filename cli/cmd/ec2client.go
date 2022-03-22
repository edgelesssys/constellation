package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/ec2/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type ec2client interface {
	GetState() (state.ConstellationState, error)
	SetState(stat state.ConstellationState) error
	CreateInstances(ctx context.Context, input client.CreateInput) error
	TerminateInstances(ctx context.Context) error
	CreateSecurityGroup(ctx context.Context, input client.SecurityGroupInput) error
	DeleteSecurityGroup(ctx context.Context) error
}
