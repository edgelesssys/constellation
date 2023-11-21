/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/terraform"
)

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}

type tfDestroyer interface {
	CleanUpWorkspace() error
	Destroy(ctx context.Context, logLevel terraform.LogLevel) error
	RemoveInstaller()
}

type tfPlanner interface {
	ShowPlan(ctx context.Context, logLevel terraform.LogLevel, output io.Writer) error
	Plan(ctx context.Context, logLevel terraform.LogLevel) (bool, error)
	PrepareWorkspace(path string, vars terraform.Variables) error
}

type tfResourceClient interface {
	tfDestroyer
	tfPlanner
	ApplyCluster(ctx context.Context, provider cloudprovider.Provider, logLevel terraform.LogLevel) (state.Infrastructure, error)
}

type tfIAMClient interface {
	tfDestroyer
	PrepareWorkspace(path string, vars terraform.Variables) error
	ApplyIAM(ctx context.Context, provider cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.IAMOutput, error)
	ShowIAM(ctx context.Context, provider cloudprovider.Provider) (terraform.IAMOutput, error)
}

type tfIAMUpgradeClient interface {
	tfPlanner
	ApplyIAM(ctx context.Context, csp cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.IAMOutput, error)
}

type libvirtRunner interface {
	Start(ctx context.Context, containerName, imageName string) error
	Stop(ctx context.Context) error
}

type rawDownloader interface {
	Download(ctx context.Context, errWriter io.Writer, isTTY bool, source, version string) (string, error)
}
