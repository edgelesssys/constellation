/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMMigrate(t *testing.T) {
	assert := assert.New(t)
	upgradeID := "test-upgrade"
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformIAMUpgradeWorkingDir)
	fs, file := setupMemFSAndFileHandler(t, []string{"terraform.tfvars", "terraform.tfstate"}, []byte("OLD"))
	csp := cloudprovider.AWS

	// act
	fakeTfClient := &tfIAMUpgradeStub{upgradeID: upgradeID, file: file}
	sut := &IAMUpgrader{
		tf:                fakeTfClient,
		logLevel:          terraform.LogLevelDebug,
		existingWorkspace: constants.TerraformIAMWorkingDir,
		upgradeWorkspace:  filepath.Join(constants.UpgradeDir, upgradeID),
		fileHandler:       file,
	}
	hasDiff, err := sut.PlanIAMUpgrade(context.Background(), io.Discard, csp)

	// assert
	assert.NoError(err)
	assert.False(hasDiff)
	assertFileExists(t, fs, filepath.Join(upgradeDir, "terraform.tfvars"))
	assertFileExists(t, fs, filepath.Join(upgradeDir, "terraform.tfstate"))

	// act
	err = sut.ApplyIAMUpgrade(context.Background(), csp)
	assert.NoError(err)

	// assert
	assertFileReadsContent(t, file, filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfvars"), "NEW")
	assertFileReadsContent(t, file, filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfstate"), "NEW")
	assertFileDoesntExist(t, fs, filepath.Join(upgradeDir))
}

func assertFileReadsContent(t *testing.T, file file.Handler, path string, expectedContent string) {
	t.Helper()
	bt, err := file.Read(path)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(bt))
}

func assertFileExists(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	res, err := fs.Stat(path)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func assertFileDoesntExist(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
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

type tfIAMUpgradeStub struct {
	upgradeID string
	file      file.Handler
	applyErr  error
	planErr   error
	planDiff  bool
	showErr   error
}

func (t *tfIAMUpgradeStub) Plan(_ context.Context, _ terraform.LogLevel) (bool, error) {
	return t.planDiff, t.planErr
}

func (t *tfIAMUpgradeStub) ShowPlan(_ context.Context, _ terraform.LogLevel, _ io.Writer) error {
	return t.showErr
}

func (t *tfIAMUpgradeStub) ApplyIAM(_ context.Context, _ cloudprovider.Provider, _ terraform.LogLevel) (terraform.IAMOutput, error) {
	if t.applyErr != nil {
		return terraform.IAMOutput{}, t.applyErr
	}

	upgradeDir := filepath.Join(constants.UpgradeDir, t.upgradeID, constants.TerraformIAMUpgradeWorkingDir)
	err := t.file.Remove(filepath.Join(upgradeDir, "terraform.tfvars"))
	if err != nil {
		return terraform.IAMOutput{}, err
	}
	err = t.file.Write(filepath.Join(upgradeDir, "terraform.tfvars"), []byte("NEW"))
	if err != nil {
		return terraform.IAMOutput{}, err
	}
	err = t.file.Remove(filepath.Join(upgradeDir, "terraform.tfstate"))
	if err != nil {
		return terraform.IAMOutput{}, err
	}
	err = t.file.Write(filepath.Join(upgradeDir, "terraform.tfstate"), []byte("NEW"))
	if err != nil {
		return terraform.IAMOutput{}, err
	}
	return terraform.IAMOutput{}, nil
}
