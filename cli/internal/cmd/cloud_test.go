package cmd

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

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
