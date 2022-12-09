/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// IAMCreator creates the IAM configuration on the cloud provider.
type IAMCreator struct {
	out                io.Writer
	newTerraformClient func(ctx context.Context) (terraformClient, error)
}

// IAMConfig holds the necessary values for IAM configuration.
type IAMConfig struct {
	GCP   GCPIAMConfig
	Azure AzureIAMConfig
	AWS   AWSIAMConfig
}

// GCPIAMConfig holds the necessary values for GCP IAM configuration.
type GCPIAMConfig struct {
	Region           string
	Zone             string
	ProjectID        string
	ServiceAccountID string
}

// AzureIAMConfig holds the necessary values for Azure IAM configuration.
type AzureIAMConfig struct {
	Region           string
	ServicePrincipal string
	ResourceGroup    string
}

// AWSIAMConfig holds the necessary values for AWS IAM configuration.
type AWSIAMConfig struct {
	Region string
	Prefix string
}

// NewIAMCreator creates a new IAM creator.
func NewIAMCreator(out io.Writer) *IAMCreator {
	return &IAMCreator{
		out: out,
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx, constants.TerraformIAMWorkingDir)
		},
	}
}

// Create prepares and hands over the corresponding providers IAM creator.
func (c *IAMCreator) Create(ctx context.Context, provider cloudprovider.Provider, iamConfig *IAMConfig) (iamid.File, error) {
	switch provider {
	case cloudprovider.GCP:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return iamid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createGCP(ctx, cl, iamConfig)
	case cloudprovider.Azure:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return iamid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAzure(ctx, cl, iamConfig)
	case cloudprovider.AWS:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return iamid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAWS(ctx, cl, iamConfig)
	default:
		return iamid.File{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// createGCP creates the IAM configuration on GCP.
func (c *IAMCreator) createGCP(ctx context.Context, cl terraformClient, iamConfig *IAMConfig) (retFile iamid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := terraform.GCPIAMVariables{
		ServiceAccountID: iamConfig.GCP.ServiceAccountID,
		Project:          iamConfig.GCP.ProjectID,
		Region:           iamConfig.GCP.Region,
		Zone:             iamConfig.GCP.Zone,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", "iam", strings.ToLower(cloudprovider.GCP.String())), &vars); err != nil {
		return iamid.File{}, err
	}

	iamOutput, err := cl.CreateIAMConfig(ctx, cloudprovider.GCP)
	if err != nil {
		return iamid.File{}, err
	}

	return iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: iamOutput.GCP.SaKey,
		},
	}, nil
}

// createAzure creates the IAM configuration on Azure.
func (c *IAMCreator) createAzure(ctx context.Context, cl terraformClient, iamConfig *IAMConfig) (retFile iamid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := terraform.AzureIAMVariables{
		Region:           iamConfig.Azure.Region,
		ResourceGroup:    iamConfig.Azure.ResourceGroup,
		ServicePrincipal: iamConfig.Azure.ServicePrincipal,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", "iam", strings.ToLower(cloudprovider.Azure.String())), &vars); err != nil {
		return iamid.File{}, err
	}

	iamOutput, err := cl.CreateIAMConfig(ctx, cloudprovider.Azure)
	if err != nil {
		return iamid.File{}, err
	}

	return iamid.File{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: iamid.AzureFile{
			ApplicationID:                iamOutput.Azure.ApplicationID,
			ApplicationClientSecretValue: iamOutput.Azure.ApplicationClientSecretValue,
			SubscriptionID:               iamOutput.Azure.SubscriptionID,
			TenantID:                     iamOutput.Azure.TenantID,
			UAMIID:                       iamOutput.Azure.UAMIID,
		},
	}, nil
}

// createAWS creates the IAM configuration on AWS.
func (c *IAMCreator) createAWS(ctx context.Context, cl terraformClient, iamConfig *IAMConfig) (retFile iamid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := terraform.AWSIAMVariables{
		Region: iamConfig.AWS.Region,
		Prefix: iamConfig.AWS.Prefix,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", "iam", strings.ToLower(cloudprovider.AWS.String())), &vars); err != nil {
		return iamid.File{}, err
	}

	iamOutput, err := cl.CreateIAMConfig(ctx, cloudprovider.AWS)
	if err != nil {
		return iamid.File{}, err
	}

	return iamid.File{
		CloudProvider: cloudprovider.AWS,
		AWSOutput: iamid.AWSFile{
			WorkerNodeInstanceProfile:   iamOutput.AWS.WorkerNodeInstanceProfile,
			ControlPlaneInstanceProfile: iamOutput.AWS.ControlPlaneInstanceProfile,
		},
	}, nil
}
