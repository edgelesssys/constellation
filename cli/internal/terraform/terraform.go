/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package terraform handles creation/destruction of a Constellation cluster using Terraform.

Since Terraform does not provide a stable Go API, we use the `terraform-exec` package to interact with Terraform.

The Terraform templates are located in the "terraform" subdirectory. The templates are embedded into the CLI binary using `go:embed`.
On use the relevant template is extracted to the working directory and the user customized variables are written to a `terraform.tfvars` file.
*/
package terraform

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
)

const (
	tfVersion         = ">= 1.2.0"
	terraformVarsFile = "terraform.tfvars"
)

// ErrTerraformWorkspaceExistsWithDifferentVariables is returned when existing Terraform files differ from the version the CLI wants to extract.
var ErrTerraformWorkspaceExistsWithDifferentVariables = errors.New("creating cluster: a Terraform workspace already exists with different variables")

// Client manages interaction with Terraform.
type Client struct {
	tf tfInterface

	file       file.Handler
	workingDir string
	remove     func()
}

// New sets up a new Client for Terraform.
func New(ctx context.Context, workingDir string) (*Client, error) {
	file := file.NewHandler(afero.NewOsFs())
	if err := file.MkdirAll(workingDir); err != nil {
		return nil, err
	}
	tf, remove, err := GetExecutable(ctx, workingDir)
	if err != nil {
		return nil, err
	}

	return &Client{
		tf:         tf,
		remove:     remove,
		file:       file,
		workingDir: workingDir,
	}, nil
}

// Show reads the default state path and outputs the state.
func (c *Client) Show(ctx context.Context) (*tfjson.State, error) {
	return c.tf.Show(ctx)
}

// PrepareWorkspace prepares a Terraform workspace for a Constellation cluster.
func (c *Client) PrepareWorkspace(path string, vars Variables) error {
	if err := prepareWorkspace(path, c.file, c.workingDir); err != nil {
		return err
	}

	if err := c.writeVars(vars); err != nil {
		return err
	}

	return nil
}

// CreateCluster creates a Constellation cluster using Terraform.
func (c *Client) CreateCluster(ctx context.Context) (CreateOutput, error) {
	if err := c.tf.Init(ctx); err != nil {
		return CreateOutput{}, err
	}

	if err := c.tf.Apply(ctx); err != nil {
		return CreateOutput{}, err
	}

	tfState, err := c.tf.Show(ctx)
	if err != nil {
		return CreateOutput{}, err
	}

	ipOutput, ok := tfState.Values.Outputs["ip"]
	if !ok {
		return CreateOutput{}, errors.New("no IP output found")
	}
	ip, ok := ipOutput.Value.(string)
	if !ok {
		return CreateOutput{}, errors.New("invalid type in IP output: not a string")
	}

	secretOutput, ok := tfState.Values.Outputs["initSecret"]
	if !ok {
		return CreateOutput{}, errors.New("no initSecret output found")
	}
	secret, ok := secretOutput.Value.(string)
	if !ok {
		return CreateOutput{}, errors.New("invalid type in initSecret output: not a string")
	}

	uidOutput, ok := tfState.Values.Outputs["uid"]
	if !ok {
		return CreateOutput{}, errors.New("no uid output found")
	}
	uid, ok := uidOutput.Value.(string)
	if !ok {
		return CreateOutput{}, errors.New("invalid type in uid output: not a string")
	}

	var attestationURL string
	if attestationURLOutput, ok := tfState.Values.Outputs["attestationURL"]; ok {
		if attestationURLString, ok := attestationURLOutput.Value.(string); ok {
			attestationURL = attestationURLString
		}
	}

	return CreateOutput{
		IP:             ip,
		Secret:         secret,
		UID:            uid,
		AttestationURL: attestationURL,
	}, nil
}

// CreateOutput contains the Terraform output values of a cluster creation.
type CreateOutput struct {
	IP     string
	Secret string
	UID    string
	// AttestationURL is the URL of the attestation provider.
	// It is only set if the cluster is created on Azure.
	AttestationURL string
}

// IAMOutput contains the output information of the Terraform IAM operations.
type IAMOutput struct {
	GCP   GCPIAMOutput
	Azure AzureIAMOutput
	AWS   AWSIAMOutput
}

// GCPIAMOutput contains the output information of the Terraform IAM operation on GCP.
type GCPIAMOutput struct {
	SaKey string
}

// AzureIAMOutput contains the output information of the Terraform IAM operation on Microsoft Azure.
type AzureIAMOutput struct {
	SubscriptionID               string
	TenantID                     string
	ApplicationID                string
	UAMIID                       string
	ApplicationClientSecretValue string
}

// AWSIAMOutput contains the output information of the Terraform IAM operation on GCP.
type AWSIAMOutput struct {
	ControlPlaneInstanceProfile string
	WorkerNodeInstanceProfile   string
}

// CreateIAMConfig creates an IAM configuration using Terraform.
func (c *Client) CreateIAMConfig(ctx context.Context, provider cloudprovider.Provider) (IAMOutput, error) {
	if err := c.tf.Init(ctx); err != nil {
		return IAMOutput{}, err
	}

	if err := c.tf.Apply(ctx); err != nil {
		return IAMOutput{}, err
	}

	tfState, err := c.tf.Show(ctx)
	if err != nil {
		return IAMOutput{}, err
	}

	switch provider {
	case cloudprovider.GCP:
		saKeyOutputRaw, ok := tfState.Values.Outputs["sa_key"]
		if !ok {
			return IAMOutput{}, errors.New("no service account key output found")
		}
		saKeyOutput, ok := saKeyOutputRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in service account key output: not a string")
		}
		return IAMOutput{
			GCP: GCPIAMOutput{
				SaKey: saKeyOutput,
			},
		}, nil
	case cloudprovider.Azure:
		subscriptionIDRaw, ok := tfState.Values.Outputs["subscription_id"]
		if !ok {
			return IAMOutput{}, errors.New("no subscription id output found")
		}
		subscriptionIDOutput, ok := subscriptionIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in subscription id output: not a string")
		}
		tenantIDRaw, ok := tfState.Values.Outputs["tenant_id"]
		if !ok {
			return IAMOutput{}, errors.New("no tenant id output found")
		}
		tenantIDOutput, ok := tenantIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in tenant id output: not a string")
		}
		applicationIDRaw, ok := tfState.Values.Outputs["application_id"]
		if !ok {
			return IAMOutput{}, errors.New("no application id output found")
		}
		applicationIDOutput, ok := applicationIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in application id output: not a string")
		}
		uamiIDRaw, ok := tfState.Values.Outputs["uami_id"]
		if !ok {
			return IAMOutput{}, errors.New("no UAMI id output found")
		}
		uamiIDOutput, ok := uamiIDRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in UAMI id output: not a string")
		}
		appClientSecretRaw, ok := tfState.Values.Outputs["application_client_secret_value"]
		if !ok {
			return IAMOutput{}, errors.New("no application client secret value output found")
		}
		appClientSecretOutput, ok := appClientSecretRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in application client secret valueoutput: not a string")
		}
		return IAMOutput{
			Azure: AzureIAMOutput{
				SubscriptionID:               subscriptionIDOutput,
				TenantID:                     tenantIDOutput,
				ApplicationID:                applicationIDOutput,
				UAMIID:                       uamiIDOutput,
				ApplicationClientSecretValue: appClientSecretOutput,
			},
		}, nil
	case cloudprovider.AWS:
		controlPlaneProfileRaw, ok := tfState.Values.Outputs["control_plane_instance_profile"]
		if !ok {
			return IAMOutput{}, errors.New("no control plane instance profile output found")
		}
		controlPlaneProfileOutput, ok := controlPlaneProfileRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in control plane instance profile output: not a string")
		}
		workerNodeProfileRaw, ok := tfState.Values.Outputs["worker_nodes_instance_profile"]
		if !ok {
			return IAMOutput{}, errors.New("no worker node instance profile output found")
		}
		workerNodeProfileOutput, ok := workerNodeProfileRaw.Value.(string)
		if !ok {
			return IAMOutput{}, errors.New("invalid type in worker node instance profile output: not a string")
		}
		return IAMOutput{
			AWS: AWSIAMOutput{
				ControlPlaneInstanceProfile: controlPlaneProfileOutput,
				WorkerNodeInstanceProfile:   workerNodeProfileOutput,
			},
		}, nil
	default:
		return IAMOutput{}, errors.New("unsupported cloud provider")
	}
}

// Destroy destroys Terraform-created cloud resources.
func (c *Client) Destroy(ctx context.Context) error {
	if err := c.tf.Init(ctx); err != nil {
		return err
	}
	return c.tf.Destroy(ctx)
}

// RemoveInstaller removes the Terraform installer, if it was downloaded for this command.
func (c *Client) RemoveInstaller() {
	c.remove()
}

// CleanUpWorkspace removes terraform files from the current directory.
func (c *Client) CleanUpWorkspace() error {
	if err := cleanUpWorkspace(c.file, c.workingDir); err != nil {
		return err
	}

	return nil
}

// GetExecutable returns a Terraform executable either from the local filesystem,
// or downloads the latest version fulfilling the version constraint.
func GetExecutable(ctx context.Context, workingDir string) (terraform *tfexec.Terraform, remove func(), err error) {
	inst := install.NewInstaller()

	version, err := version.NewConstraint(tfVersion)
	if err != nil {
		return nil, nil, err
	}

	downloadVersion := &releases.LatestVersion{
		Product:     product.Terraform,
		Constraints: version,
	}
	localVersion := &fs.Version{
		Product:     product.Terraform,
		Constraints: version,
	}

	execPath, err := inst.Ensure(ctx, []src.Source{localVersion, downloadVersion})
	if err != nil {
		return nil, nil, err
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)

	return tf, func() { _ = inst.Remove(context.Background()) }, err
}

// writeVars tries to write the Terraform variables file or, if it exists, checks if it is the same as we are expecting.
func (c *Client) writeVars(vars Variables) error {
	if vars == nil {
		return errors.New("creating cluster: vars is nil")
	}

	pathToVarsFile := filepath.Join(c.workingDir, terraformVarsFile)
	if err := c.file.Write(pathToVarsFile, []byte(vars.String())); errors.Is(err, afero.ErrFileExists) {
		// If a variables file already exists, check if it's the same as we're expecting, so we can continue using it.
		varsContent, err := c.file.Read(pathToVarsFile)
		if err != nil {
			return err
		}
		if vars.String() != string(varsContent) {
			return ErrTerraformWorkspaceExistsWithDifferentVariables
		}
	} else if err != nil {
		return err
	}

	return nil
}

type tfInterface interface {
	Apply(context.Context, ...tfexec.ApplyOption) error
	Destroy(context.Context, ...tfexec.DestroyOption) error
	Init(context.Context, ...tfexec.InitOption) error
	Show(context.Context, ...tfexec.ShowOption) (*tfjson.State, error)
}
