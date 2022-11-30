/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
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

// PrepareWorkspace prepares a Terraform workspace for a Constellation cluster.
func (c *Client) PrepareWorkspace(provider cloudprovider.Provider, vars Variables) error {
	if err := prepareWorkspace(c.file, provider, c.workingDir); err != nil {
		return err
	}

	if err := c.writeVars(vars); err != nil {
		return err
	}

	return nil
}

// CreateCluster creates a Constellation cluster using Terraform.
func (c *Client) CreateCluster(ctx context.Context) (string, string, error) {
	if err := c.tf.Init(ctx); err != nil {
		return "", "", err
	}

	if err := c.tf.Apply(ctx); err != nil {
		return "", "", err
	}

	tfState, err := c.tf.Show(ctx)
	if err != nil {
		return "", "", err
	}

	ipOutput, ok := tfState.Values.Outputs["ip"]
	if !ok {
		return "", "", errors.New("no IP output found")
	}
	ip, ok := ipOutput.Value.(string)
	if !ok {
		return "", "", errors.New("invalid type in IP output: not a string")
	}

	secretOutput, ok := tfState.Values.Outputs["initSecret"]
	if !ok {
		return "", "", errors.New("no initSecret output found")
	}
	secret, ok := secretOutput.Value.(string)
	if !ok {
		return "", "", errors.New("invalid type in initSecret output: not a string")
	}

	return ip, secret, nil
}

// DestroyCluster destroys a Constellation cluster using Terraform.
func (c *Client) DestroyCluster(ctx context.Context) error {
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
