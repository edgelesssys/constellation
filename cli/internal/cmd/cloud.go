/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
)

type cloudCreator interface {
	Create(
		ctx context.Context,
		opts cloudcmd.CreateOptions,
	) (clusterid.File, error)
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
