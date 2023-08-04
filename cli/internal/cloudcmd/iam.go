/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
)

// IAMDestroyer destroys an IAM configuration.
type IAMDestroyer struct {
	newTerraformClient newTFIAMClientFunc
}

// NewIAMDestroyer creates a new IAM Destroyer.
func NewIAMDestroyer() *IAMDestroyer {
	return &IAMDestroyer{newTerraformClient: newTerraformIAMClient}
}

// GetTfStateServiceAccountKey returns the sa_key output from the terraform state.
func (d *IAMDestroyer) GetTfStateServiceAccountKey(ctx context.Context, tfWorkspace string) (gcpshared.ServiceAccountKey, error) {
	client, err := d.newTerraformClient(ctx, tfWorkspace)
	if err != nil {
		return gcpshared.ServiceAccountKey{}, err
	}

	tfState, err := client.ShowIAM(ctx, cloudprovider.GCP)
	if err != nil {
		return gcpshared.ServiceAccountKey{}, fmt.Errorf("getting terraform state: %w", err)
	}
	if saKeyString := tfState.GCP.SaKey; saKeyString != "" {
		saKey, err := base64.StdEncoding.DecodeString(saKeyString)
		if err != nil {
			return gcpshared.ServiceAccountKey{}, err
		}
		var tfSaKey gcpshared.ServiceAccountKey
		if err := json.Unmarshal(saKey, &tfSaKey); err != nil {
			return gcpshared.ServiceAccountKey{}, err
		}

		return tfSaKey, nil
	}
	return gcpshared.ServiceAccountKey{}, errors.New("no saKey in terraform state")
}

// DestroyIAMConfiguration destroys the previously created IAM configuration and deletes the local IAM terraform files.
func (d *IAMDestroyer) DestroyIAMConfiguration(ctx context.Context, tfWorkspace string, logLevel terraform.LogLevel) error {
	client, err := d.newTerraformClient(ctx, tfWorkspace)
	if err != nil {
		return err
	}

	if err := client.Destroy(ctx, logLevel); err != nil {
		return err
	}
	return client.CleanUpWorkspace()
}

// IAMCreator creates the IAM configuration on the cloud provider.
type IAMCreator struct {
	out                io.Writer
	newTerraformClient newTFIAMClientFunc
}

// IAMConfigOptions holds the necessary values for IAM configuration.
type IAMConfigOptions struct {
	GCP         GCPIAMConfig
	Azure       AzureIAMConfig
	AWS         AWSIAMConfig
	TFLogLevel  terraform.LogLevel
	TFWorkspace string
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
		out:                out,
		newTerraformClient: newTerraformIAMClient,
	}
}

// Create prepares and hands over the corresponding providers IAM creator.
func (c *IAMCreator) Create(ctx context.Context, provider cloudprovider.Provider, opts *IAMConfigOptions) (iamid.File, error) {
	cl, err := c.newTerraformClient(ctx, opts.TFWorkspace)
	if err != nil {
		return iamid.File{}, err
	}
	defer cl.RemoveInstaller()

	switch provider {
	case cloudprovider.GCP:
		return c.createGCP(ctx, cl, opts)
	case cloudprovider.Azure:
		return c.createAzure(ctx, cl, opts)
	case cloudprovider.AWS:
		return c.createAWS(ctx, cl, opts)
	default:
		return iamid.File{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// createGCP creates the IAM configuration on GCP.
func (c *IAMCreator) createGCP(ctx context.Context, cl tfIAMClient, opts *IAMConfigOptions) (retFile iamid.File, retErr error) {
	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)

	vars := terraform.GCPIAMVariables{
		ServiceAccountID: opts.GCP.ServiceAccountID,
		Project:          opts.GCP.ProjectID,
		Region:           opts.GCP.Region,
		Zone:             opts.GCP.Zone,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", "iam", strings.ToLower(cloudprovider.GCP.String())), &vars); err != nil {
		return iamid.File{}, err
	}

	iamOutput, err := cl.ApplyIAMConfig(ctx, cloudprovider.GCP, opts.TFLogLevel)
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
func (c *IAMCreator) createAzure(ctx context.Context, cl tfIAMClient, opts *IAMConfigOptions) (retFile iamid.File, retErr error) {
	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)

	vars := terraform.AzureIAMVariables{
		Region:           opts.Azure.Region,
		ResourceGroup:    opts.Azure.ResourceGroup,
		ServicePrincipal: opts.Azure.ServicePrincipal,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", "iam", strings.ToLower(cloudprovider.Azure.String())), &vars); err != nil {
		return iamid.File{}, err
	}

	iamOutput, err := cl.ApplyIAMConfig(ctx, cloudprovider.Azure, opts.TFLogLevel)
	if err != nil {
		return iamid.File{}, err
	}

	return iamid.File{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: iamid.AzureFile{
			SubscriptionID: iamOutput.Azure.SubscriptionID,
			TenantID:       iamOutput.Azure.TenantID,
			UAMIID:         iamOutput.Azure.UAMIID,
		},
	}, nil
}

// createAWS creates the IAM configuration on AWS.
func (c *IAMCreator) createAWS(ctx context.Context, cl tfIAMClient, opts *IAMConfigOptions) (retFile iamid.File, retErr error) {
	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)

	vars := terraform.AWSIAMVariables{
		Region: opts.AWS.Region,
		Prefix: opts.AWS.Prefix,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", "iam", strings.ToLower(cloudprovider.AWS.String())), &vars); err != nil {
		return iamid.File{}, err
	}

	iamOutput, err := cl.ApplyIAMConfig(ctx, cloudprovider.AWS, opts.TFLogLevel)
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

type newTFIAMClientFunc func(ctx context.Context, workspace string) (tfIAMClient, error)

func newTerraformIAMClient(ctx context.Context, workspace string) (tfIAMClient, error) {
	return terraform.New(ctx, workspace)
}
