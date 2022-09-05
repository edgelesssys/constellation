/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
)

type cloudCreator interface {
	Create(
		ctx context.Context,
		provider cloudprovider.Provider,
		config *config.Config,
		name, insType string,
		coordCount, nodeCount int,
	) (state.ConstellationState, error)
}

type cloudTerminator interface {
	Terminate(context.Context, state.ConstellationState) error
}
