/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
)

type cloudApplier interface {
	Plan(ctx context.Context, conf *config.Config) (bool, error)
	Apply(ctx context.Context, csp cloudprovider.Provider, rollback cloudcmd.RollbackBehavior) (state.Infrastructure, error)
	RestoreWorkspace() error
	WorkingDirIsEmpty() (bool, error)
}

type cloudIAMCreator interface {
	Create(
		ctx context.Context,
		provider cloudprovider.Provider,
		opts *cloudcmd.IAMConfigOptions,
	) (cloudcmd.IAMOutput, error)
}

type iamDestroyer interface {
	DestroyIAMConfiguration(ctx context.Context, tfWorkspace string, logLevel terraform.LogLevel) error
	GetTfStateServiceAccountKey(ctx context.Context, tfWorkspace string) (gcpshared.ServiceAccountKey, error)
}

type cloudTerminator interface {
	Terminate(ctx context.Context, workspace string, logLevel terraform.LogLevel) error
}
