package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func defaultConfigAsYAML(t *testing.T) string {
	var readBuffer bytes.Buffer
	require.NoError(t, yaml.NewEncoder(&readBuffer).Encode(config.Default()))
	return readBuffer.String()
}

func TestConfigGenerateDefault(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()

	require.NoError(configGenerate(cmd, fileHandler))

	readYAML, err := fileHandler.Read(constants.ConfigFilename)
	assert.NoError(err)
	assert.Equal(defaultConfigAsYAML(t), string(readYAML))
}

func TestConfigGenerateDefaultExists(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	require.NoError(fileHandler.Write(constants.ConfigFilename, []byte("foobar"), file.OptNone))
	cmd := newConfigGenerateCmd()

	require.Error(configGenerate(cmd, fileHandler))
}

func TestConfigGenerateFileFlagRemoved(t *testing.T) {
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())
	cmd := newConfigGenerateCmd()
	cmd.ResetFlags()

	require.Error(configGenerate(cmd, fileHandler))
}

func TestConfigGenerateStdOut(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fileHandler := file.NewHandler(afero.NewMemMapFs())

	var outBuffer bytes.Buffer
	cmd := newConfigGenerateCmd()
	cmd.SetOut(&outBuffer)
	require.NoError(cmd.Flags().Set("file", "-"))

	require.NoError(configGenerate(cmd, fileHandler))

	assert.Equal(defaultConfigAsYAML(t), outBuffer.String())
}
