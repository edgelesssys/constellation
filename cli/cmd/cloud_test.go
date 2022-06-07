package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

type stubCloudCreator struct {
	createCalled bool
	state        state.ConstellationState
	createErr    error
}

func (c *stubCloudCreator) Create(
	ctx context.Context,
	provider cloudprovider.Provider,
	config *config.Config,
	name, insType string,
	coordCount, nodeCount int,
) (state.ConstellationState, error) {
	c.createCalled = true
	return c.state, c.createErr
}

type stubCloudTerminator struct {
	called       bool
	terminateErr error
}

func (c *stubCloudTerminator) Terminate(context.Context, state.ConstellationState) error {
	c.called = true
	return c.terminateErr
}

func (c *stubCloudTerminator) Called() bool {
	return c.called
}

type stubServiceAccountCreator struct {
	cloudServiceAccountURI string
	createErr              error
}

func (c *stubServiceAccountCreator) Create(ctx context.Context, stat state.ConstellationState, config *config.Config) (string, state.ConstellationState, error) {
	return c.cloudServiceAccountURI, stat, c.createErr
}
