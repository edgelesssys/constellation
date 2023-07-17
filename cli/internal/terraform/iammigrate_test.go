/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMMigrate(t *testing.T) {
	// plan should copy the terraform files to the upgrade directory
	// create upgrade dir
	// show plan changes in outWriter
	upgradeID := "test-upgrade"
	tfClient, err := New(context.Background(), filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeWorkingDir))
	require.NoError(t, err)
	// prep fs
	fs := afero.NewMemMapFs()
	file := file.NewHandler(fs)
	err = file.MkdirAll(constants.TerraformIAMWorkingDir)
	require.NoError(t, err)
	err = file.Write(filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfvars"), []byte("OLD"))
	require.NoError(t, err)
	err = file.Write(filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfstate"), []byte("OLD"))
	require.NoError(t, err)
	tfClient.file = file

	writer := bytes.NewBuffer(nil)
	fakeTfClient := &tfClientStub{tfClient, upgradeID}
	sut := NewIAMMigrateCmd(fakeTfClient, upgradeID, cloudprovider.AWS, LogLevelDebug, writer)

	hasDiff, err := sut.Plan(context.Background())
	assert.NoError(t, err)
	assert.False(t, hasDiff)
	// check that files are copied
	res, err := fs.Stat(filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformIAMUpgradeWorkingDir, "terraform.tfvars"))

	assert.NoError(t, err)
	assert.NotNil(t, res)
	res, err = fs.Stat(filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformIAMUpgradeWorkingDir, "terraform.tfstate"))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// apply
	err = sut.Apply(context.Background(), file)
	assert.NoError(t, err)
	// check that files are copied
	bt, err := file.Read(filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfvars"))
	assert.NoError(t, err)
	assert.Equal(t, "NEW", string(bt))
	bt, err = file.Read(filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfstate"))
	assert.NoError(t, err)
	assert.Equal(t, "NEW", string(bt))

	// upgrade dir should be removed
	res, err = fs.Stat(filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformIAMUpgradeWorkingDir))
	assert.Error(t, err)
	assert.Nil(t, res)
}

type tfClientStub struct {
	realClient *Client
	upgradeID  string
}

func (t *tfClientStub) PrepareIAMUpgradeWorkspace(rootDir, workingDir, newWorkingDir, backupDir string) error {
	return t.realClient.PrepareIAMUpgradeWorkspace(rootDir, workingDir, newWorkingDir, backupDir)
}

func (t *tfClientStub) Plan(ctx context.Context, logLevel LogLevel, planFile string) (bool, error) {
	return false, nil
}

func (t *tfClientStub) ShowPlan(ctx context.Context, logLevel LogLevel, planFile string, outWriter io.Writer) error {
	return nil
}

func (t *tfClientStub) CreateIAMConfig(ctx context.Context, csp cloudprovider.Provider, logLevel LogLevel) (IAMOutput, error) {
	upgradeDir := filepath.Join(constants.UpgradeDir, t.upgradeID, constants.TerraformIAMUpgradeWorkingDir)
	err := t.realClient.file.Remove(filepath.Join(upgradeDir, "terraform.tfvars"))
	if err != nil {
		return IAMOutput{}, err
	}
	err = t.realClient.file.Write(filepath.Join(upgradeDir, "terraform.tfvars"), []byte("NEW"))
	if err != nil {
		return IAMOutput{}, err
	}
	err = t.realClient.file.Remove(filepath.Join(upgradeDir, "terraform.tfstate"))
	if err != nil {
		return IAMOutput{}, err
	}
	err = t.realClient.file.Write(filepath.Join(upgradeDir, "terraform.tfstate"), []byte("NEW"))
	if err != nil {
		return IAMOutput{}, err
	}
	return IAMOutput{}, nil
}
