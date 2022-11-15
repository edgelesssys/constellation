/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

type terraformClient interface {
	CreateCluster(ctx context.Context, provider cloudprovider.Provider, input terraform.Variables) (string, error)
	DestroyCluster(ctx context.Context) error
	CleanUpWorkspace() error
	RemoveInstaller()
}

type libvirtRunner interface {
	Start(ctx context.Context, containerName, imageName string) error
	Stop(ctx context.Context) error
}
