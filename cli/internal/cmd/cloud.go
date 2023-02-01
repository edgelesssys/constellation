/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
)

type cloudCreator interface {
	Create(
		ctx context.Context,
		provider cloudprovider.Provider,
		config *config.Config,
		insType string,
		coordCount, nodeCount int,
	) (clusterid.File, error)
}

type cloudIAMCreator interface {
	Create(
		ctx context.Context,
		provider cloudprovider.Provider,
		iamConfig *cloudcmd.IAMConfig,
	) (iamid.File, error)
}

type iamDestroyer interface {
	DestroyIAMConfiguration(ctx context.Context) error
	RunDeleteGCPKeyFile(ctx context.Context) (bool, error)
}

type cloudTerminator interface {
	Terminate(context.Context) error
}
