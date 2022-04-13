package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

type cloudCreator interface {
	Create(
		ctx context.Context,
		provider cloudprovider.CloudProvider,
		config *config.Config,
		name, insType string,
		coordCount, nodeCount int,
	) (state.ConstellationState, error)
}

type cloudTerminator interface {
	Terminate(context.Context, state.ConstellationState) error
}

type serviceAccountCreator interface {
	Create(ctx context.Context, stat state.ConstellationState, config *config.Config,
	) (string, state.ConstellationState, error)
}
