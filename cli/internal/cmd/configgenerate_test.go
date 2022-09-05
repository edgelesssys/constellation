/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfigGenerateDefault(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()

	require.NoError(configGenerate(cmd, fileHandler, cloudprovider.Unknown))

	var readConfig config.Config
	err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
	assert.NoError(err)
	assert.Equal(*config.Default(), readConfig)
}

func TestConfigGenerateDefaultGCPSpecific(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	wantConf := config.Default()
	wantConf.RemoveProviderExcept(cloudprovider.GCP)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()

	require.NoError(configGenerate(cmd, fileHandler, cloudprovider.GCP))

	var readConfig config.Config
	err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
	assert.NoError(err)
	assert.Equal(*wantConf, readConfig)
}

func TestConfigGenerateDefaultExists(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	require.NoError(fileHandler.Write(constants.ConfigFilename, []byte("foobar"), file.OptNone))
	cmd := newConfigGenerateCmd()

	require.Error(configGenerate(cmd, fileHandler, cloudprovider.Unknown))
}

func TestConfigGenerateFileFlagRemoved(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()
	cmd.ResetFlags()

	require.Error(configGenerate(cmd, fileHandler, cloudprovider.Unknown))
}

func TestConfigGenerateStdOut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())

	var outBuffer bytes.Buffer
	cmd := newConfigGenerateCmd()
	cmd.SetOut(&outBuffer)
	require.NoError(cmd.Flags().Set("file", "-"))

	require.NoError(configGenerate(cmd, fileHandler, cloudprovider.Unknown))

	var readConfig config.Config
	require.NoError(yaml.NewDecoder(&outBuffer).Decode(&readConfig))

	assert.Equal(*config.Default(), readConfig)
}
