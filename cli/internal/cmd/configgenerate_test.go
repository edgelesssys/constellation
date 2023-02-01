/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfigGenerateDefault(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()
	nameFlag := "constell"
	require.NoError(cmd.Flags().Set("name", nameFlag))

	wantConf := config.Default()
	wantConf.Name = nameFlag

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown))

	var readConfig config.Config
	err := fileHandler.ReadYAML(constants.ConfigFilename, &readConfig)
	assert.NoError(err)
	assert.Equal(*wantConf, readConfig)
}

func TestConfigGenerateDefaultGCPSpecific(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()
	nameFlag := "constell"
	require.NoError(cmd.Flags().Set("name", nameFlag))

	wantConf := config.Default()
	wantConf.RemoveProviderExcept(cloudprovider.GCP)
	wantConf.Name = nameFlag

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.GCP))

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

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.Error(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown))
}

func TestConfigGenerateFileFlagRemoved(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()
	cmd.ResetFlags()

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.Error(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown))
}

func TestConfigGenerateStdOut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())

	var outBuffer bytes.Buffer
	cmd := newConfigGenerateCmd()
	cmd.SetOut(&outBuffer)
	require.NoError(cmd.Flags().Set("file", "-"))
	nameFlag := "constell"
	require.NoError(cmd.Flags().Set("name", nameFlag))

	cg := &configGenerateCmd{log: logger.NewTest(t)}
	require.NoError(cg.configGenerate(cmd, fileHandler, cloudprovider.Unknown))

	var readConfig config.Config
	require.NoError(yaml.NewDecoder(&outBuffer).Decode(&readConfig))

	wantConfig := config.Default()
	wantConfig.Name = nameFlag
	assert.Equal(*wantConfig, readConfig)
}

func TestParseNameFlag(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	wantName := "constell"
	cmd := &cobra.Command{}
	cmd.Flags().String("name", wantName, "")

	// default name flag value
	name, err := parseNameFlag(cmd)
	require.NoError(err)
	assert.Equal(wantName, name)

	wantName = "kubernetes"
	require.NoError(cmd.Flags().Set("name", wantName))
	name, err = parseNameFlag(cmd)
	require.NoError(err)
	assert.Equal(wantName, name)

	require.NoError(cmd.Flags().Set("name", strings.Repeat("a", constants.ConstellationNameLength+1)))
	_, err = parseNameFlag(cmd)
	assert.Error(err)
}
