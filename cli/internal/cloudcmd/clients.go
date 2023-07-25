/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}

type tfCommonClient interface {
	CleanUpWorkspace() error
	Destroy(ctx context.Context, logLevel terraform.LogLevel) error
	PrepareWorkspace(path string, input terraform.Variables) error
	RemoveInstaller()
}

type tfResourceClient interface {
	tfCommonClient
	CreateCluster(ctx context.Context, logLevel terraform.LogLevel) (terraform.ApplyOutput, error)
	ShowCluster(ctx context.Context) (terraform.ApplyOutput, error)
}

type tfIAMClient interface {
	tfCommonClient
	ApplyIAMConfig(ctx context.Context, provider cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.IAMOutput, error)
	ShowIAM(ctx context.Context, provider cloudprovider.Provider) (terraform.IAMOutput, error)
}

type libvirtRunner interface {
	Start(ctx context.Context, containerName, imageName string) error
	Stop(ctx context.Context) error
}

type rawDownloader interface {
	Download(ctx context.Context, errWriter io.Writer, isTTY bool, source, version string) (string, error)
}
