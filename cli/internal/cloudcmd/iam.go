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

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// IAMDestroyer destroys an IAM configuration.
type IAMDestroyer struct {
	newTerraformClient newTFIAMClientFunc
}

// NewIAMDestroyer creates a new IAM Destroyer.
func NewIAMDestroyer() *IAMDestroyer {
	return &IAMDestroyer{newTerraformClient: newTerraformIAMClient}
}

// GetTfStateServiceAccountKey returns the service_account_key output from the terraform state.
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
	SubscriptionID   string
	Location         string
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
func (c *IAMCreator) Create(ctx context.Context, provider cloudprovider.Provider, opts *IAMConfigOptions) (IAMOutput, error) {
	cl, err := c.newTerraformClient(ctx, opts.TFWorkspace)
	if err != nil {
		return IAMOutput{}, err
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
		return IAMOutput{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// createGCP creates the IAM configuration on GCP.
func (c *IAMCreator) createGCP(ctx context.Context, cl tfIAMClient, opts *IAMConfigOptions) (iam IAMOutput, retErr error) {
	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)

	vars := terraform.GCPIAMVariables{
		ServiceAccountID: opts.GCP.ServiceAccountID,
		Project:          opts.GCP.ProjectID,
		Region:           opts.GCP.Region,
		Zone:             opts.GCP.Zone,
	}

	if err := cl.PrepareWorkspace(path.Join(constants.TerraformEmbeddedDir, "iam", strings.ToLower(cloudprovider.GCP.String())), &vars); err != nil {
		return IAMOutput{}, err
	}

	iamOutput, err := cl.ApplyIAM(ctx, cloudprovider.GCP, opts.TFLogLevel)
	if err != nil {
		return IAMOutput{}, err
	}

	return IAMOutput{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: GCPIAMOutput{
			ServiceAccountKey: iamOutput.GCP.SaKey,
		},
	}, nil
}

// createAzure creates the IAM configuration on Azure.
func (c *IAMCreator) createAzure(ctx context.Context, cl tfIAMClient, opts *IAMConfigOptions) (iam IAMOutput, retErr error) {
	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)

	vars := terraform.AzureIAMVariables{
		SubscriptionID:   opts.Azure.SubscriptionID,
		Location:         opts.Azure.Location,
		ResourceGroup:    opts.Azure.ResourceGroup,
		ServicePrincipal: opts.Azure.ServicePrincipal,
	}

	if err := cl.PrepareWorkspace(path.Join(constants.TerraformEmbeddedDir, "iam", strings.ToLower(cloudprovider.Azure.String())), &vars); err != nil {
		return IAMOutput{}, err
	}

	iamOutput, err := cl.ApplyIAM(ctx, cloudprovider.Azure, opts.TFLogLevel)
	if err != nil {
		return IAMOutput{}, err
	}

	return IAMOutput{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: AzureIAMOutput{
			SubscriptionID: iamOutput.Azure.SubscriptionID,
			TenantID:       iamOutput.Azure.TenantID,
			UAMIID:         iamOutput.Azure.UAMIID,
		},
	}, nil
}

// createAWS creates the IAM configuration on AWS.
func (c *IAMCreator) createAWS(ctx context.Context, cl tfIAMClient, opts *IAMConfigOptions) (iam IAMOutput, retErr error) {
	defer rollbackOnError(c.out, &retErr, &rollbackerTerraform{client: cl}, opts.TFLogLevel)

	vars := terraform.AWSIAMVariables{
		Region: opts.AWS.Region,
		Prefix: opts.AWS.Prefix,
	}

	if err := cl.PrepareWorkspace(path.Join(constants.TerraformEmbeddedDir, "iam", strings.ToLower(cloudprovider.AWS.String())), &vars); err != nil {
		return IAMOutput{}, err
	}

	iamOutput, err := cl.ApplyIAM(ctx, cloudprovider.AWS, opts.TFLogLevel)
	if err != nil {
		return IAMOutput{}, err
	}

	return IAMOutput{
		CloudProvider: cloudprovider.AWS,
		AWSOutput: AWSIAMOutput{
			WorkerNodeInstanceProfile:   iamOutput.AWS.WorkerNodeInstanceProfile,
			ControlPlaneInstanceProfile: iamOutput.AWS.ControlPlaneInstanceProfile,
		},
	}, nil
}

// IAMOutput is the output of creating a new IAM profile.
type IAMOutput struct {
	// CloudProvider is the cloud provider of the cluster.
	CloudProvider cloudprovider.Provider `json:"cloudprovider,omitempty"`

	GCPOutput   GCPIAMOutput   `json:"gcpOutput,omitempty"`
	AzureOutput AzureIAMOutput `json:"azureOutput,omitempty"`
	AWSOutput   AWSIAMOutput   `json:"awsOutput,omitempty"`
}

// GCPIAMOutput contains the output information of a GCP IAM configuration.
type GCPIAMOutput struct {
	ServiceAccountKey string `json:"serviceAccountID,omitempty"`
}

// AzureIAMOutput contains the output information of a Microsoft Azure IAM configuration.
type AzureIAMOutput struct {
	SubscriptionID string `json:"subscriptionID,omitempty"`
	TenantID       string `json:"tenantID,omitempty"`
	UAMIID         string `json:"uamiID,omitempty"`
}

// AWSIAMOutput contains the output information of an AWS IAM configuration.
type AWSIAMOutput struct {
	ControlPlaneInstanceProfile string `json:"controlPlaneInstanceProfile,omitempty"`
	WorkerNodeInstanceProfile   string `json:"workerNodeInstanceProfile,omitempty"`
}

type newTFIAMClientFunc func(ctx context.Context, workspace string) (tfIAMClient, error)

func newTerraformIAMClient(ctx context.Context, workspace string) (tfIAMClient, error) {
	return terraform.New(ctx, workspace)
}
