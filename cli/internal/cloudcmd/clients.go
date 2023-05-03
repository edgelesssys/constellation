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
	tfjson "github.com/hashicorp/terraform-json"
)

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}

type terraformClient interface {
	PrepareWorkspace(path string, input terraform.Variables) error
	CreateCluster(ctx context.Context, logLevel terraform.LogLevel, targets ...string) (terraform.CreateOutput, error)
	CreateIAMConfig(ctx context.Context, provider cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.IAMOutput, error)
	Destroy(ctx context.Context, logLevel terraform.LogLevel) error
	CleanUpWorkspace() error
	RemoveInstaller()
	Show(ctx context.Context) (*tfjson.State, error)
}

type libvirtRunner interface {
	Start(ctx context.Context, containerName, imageName string) error
	Stop(ctx context.Context) error
}

type rawDownloader interface {
	Download(ctx context.Context, errWriter io.Writer, isTTY bool, source, version string) (string, error)
}
