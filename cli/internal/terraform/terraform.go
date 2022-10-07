/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/state"
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

// Client manages interaction with Terraform.
type Client struct {
	tf tfInterface

	provider cloudprovider.Provider

	file   file.Handler
	state  state.ConstellationState
	remove func()
}

// New sets up a new Client for Terraform.
func New(ctx context.Context, provider cloudprovider.Provider) (*Client, error) {
	tf, remove, err := GetExecutable(ctx, ".")
	if err != nil {
		return nil, err
	}

	file := file.NewHandler(afero.NewOsFs())

	return &Client{
		tf:       tf,
		provider: provider,
		remove:   remove,
		file:     file,
	}, nil
}

// CreateCluster creates a Constellation cluster using Terraform.
func (c *Client) CreateCluster(ctx context.Context, name string, vars Variables) error {
	if err := prepareWorkspace(c.file, c.provider); err != nil {
		return err
	}

	if err := c.tf.Init(ctx); err != nil {
		return err
	}

	if err := c.file.Write(terraformVarsFile, []byte(vars.String())); err != nil {
		return err
	}

	if err := c.tf.Apply(ctx); err != nil {
		return err
	}

	tfState, err := c.tf.Show(ctx)
	if err != nil {
		return err
	}

	ipOutput, ok := tfState.Values.Outputs["ip"]
	if !ok {
		return errors.New("no IP output found")
	}
	ip, ok := ipOutput.Value.(string)
	if !ok {
		return errors.New("invalid type in IP output: not a string")
	}
	c.state = state.ConstellationState{
		Name:           name,
		CloudProvider:  c.provider.String(),
		LoadBalancerIP: ip,
	}

	return nil
}

// DestroyInstances destroys a Constellation cluster using Terraform.
func (c *Client) DestroyCluster(ctx context.Context) error {
	return c.tf.Destroy(ctx)
}

// RemoveInstaller removes the Terraform installer, if it was downloaded for this command.
func (c *Client) RemoveInstaller() {
	c.remove()
}

// CleanUpWorkspace removes terraform files from the current directory.
func (c *Client) CleanUpWorkspace() error {
	if err := cleanUpWorkspace(c.file, c.provider); err != nil {
		return err
	}

	if err := ignoreFileNotFoundErr(c.file.Remove("terraform.tfvars")); err != nil {
		return err
	}
	if err := ignoreFileNotFoundErr(c.file.Remove("terraform.tfstate")); err != nil {
		return err
	}
	if err := ignoreFileNotFoundErr(c.file.Remove("terraform.tfstate.backup")); err != nil {
		return err
	}
	if err := ignoreFileNotFoundErr(c.file.Remove(".terraform.lock.hcl")); err != nil {
		return err
	}
	if err := ignoreFileNotFoundErr(c.file.RemoveAll(".terraform")); err != nil {
		return err
	}

	return nil
}

// GetState returns the state of the cluster.
func (c *Client) GetState() state.ConstellationState {
	return c.state
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

type tfInterface interface {
	Apply(context.Context, ...tfexec.ApplyOption) error
	Destroy(context.Context, ...tfexec.DestroyOption) error
	Init(context.Context, ...tfexec.InitOption) error
	Show(context.Context, ...tfexec.ShowOption) (*tfjson.State, error)
}
