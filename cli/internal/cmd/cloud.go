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
	) (iamid.File, error)
}

type iamDestroyer interface {
	DestroyIAMConfiguration(ctx context.Context, logLevel terraform.LogLevel) error
	GetTfstateServiceAccountKey(ctx context.Context) (gcpshared.ServiceAccountKey, error)
}

type cloudTerminator interface {
	Terminate(ctx context.Context, logLevel terraform.LogLevel) error
}
