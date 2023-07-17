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
	upgradeID := "test-upgrade"
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformIAMUpgradeWorkingDir)
	tfClient, err := New(context.Background(), filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeWorkingDir))
	require.NoError(t, err)
	fs, file := setupMemFSAndFileHandler(t, []string{"terraform.tfvars", "terraform.tfstate"}, []byte("OLD"))
	tfClient.file = file

	// act
	fakeTfClient := &tfClientStub{tfClient, upgradeID}
	sut := NewIAMMigrateCmd(fakeTfClient, upgradeID, cloudprovider.AWS, LogLevelDebug, bytes.NewBuffer(nil))
	hasDiff, err := sut.Plan(context.Background())
	// assert
	assert.NoError(t, err)
	assert.False(t, hasDiff)
	assertFileExists(fs, filepath.Join(upgradeDir, "terraform.tfvars"), t)
	assertFileExists(fs, filepath.Join(upgradeDir, "terraform.tfstate"), t)

	// act
	err = sut.Apply(context.Background(), file)
	assert.NoError(t, err)
	// assert
	assertFileReadsContent(file, filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfvars"), "NEW", t)
	assertFileReadsContent(file, filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfstate"), "NEW", t)
	assertFileDoesntExist(fs, filepath.Join(upgradeDir), t)
}

func assertFileReadsContent(file file.Handler, path string, expectedContent string, t *testing.T) {
	bt, err := file.Read(path)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(bt))
}

func assertFileExists(fs afero.Fs, path string, t *testing.T) {
	res, err := fs.Stat(path)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func assertFileDoesntExist(fs afero.Fs, path string, t *testing.T) {
	res, err := fs.Stat(path)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// setupMemFSAndFileHandler sets up a file handler with a memory file system and writes the given files with the given content.
func setupMemFSAndFileHandler(t *testing.T, files []string, content []byte) (afero.Fs, file.Handler) {
	fs := afero.NewMemMapFs()
	file := file.NewHandler(fs)
	err := file.MkdirAll(constants.TerraformIAMWorkingDir)
	require.NoError(t, err)

	for _, f := range files {
		err := file.Write(filepath.Join(constants.TerraformIAMWorkingDir, f), content)
		require.NoError(t, err)
	}
	return fs, file
}

type tfClientStub struct {
	realClient *Client
	upgradeID  string
}

func (t *tfClientStub) PrepareIAMUpgradeWorkspace(rootDir, workingDir, newWorkingDir, backupDir string) error {
	return t.realClient.PrepareIAMUpgradeWorkspace(rootDir, workingDir, newWorkingDir, backupDir)
}

func (t *tfClientStub) Plan(_ context.Context, _ LogLevel, _ string) (bool, error) {
	return false, nil
}

func (t *tfClientStub) ShowPlan(_ context.Context, _ LogLevel, _ string, _ io.Writer) error {
	return nil
}

func (t *tfClientStub) CreateIAMConfig(_ context.Context, _ cloudprovider.Provider, _ LogLevel) (IAMOutput, error) {
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
