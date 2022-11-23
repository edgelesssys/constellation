/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
)

type terraformClient interface {
	PrepareWorkspace(provider cloudprovider.Provider, input terraform.Variables) error
	CreateCluster(ctx context.Context) (string, error)
	DestroyCluster(ctx context.Context) error
	CleanUpWorkspace() error
	RemoveInstaller()
}

type libvirtRunner interface {
	Start(ctx context.Context, containerName, imageName string) error
	Stop(ctx context.Context) error
}

type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}

type rawDownloader interface {
	Download(ctx context.Context, errWriter io.Writer, isTTY bool, source, version string) (string, error)
}
