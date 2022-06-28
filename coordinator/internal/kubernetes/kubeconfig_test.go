package kubernetes

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadKubeconfig(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	fs := afero.Afero{
		Fs: afero.NewMemMapFs(),
	}
	require.NoError(fs.WriteFile(kubeconfigPath, []byte("someConfig"), 0o644))
	reader := KubeconfigReader{fs}
	config, err := reader.ReadKubeconfig()

	require.NoError(err)
	assert.Equal([]byte("someConfig"), config)
}

func TestReadKubeconfigFails(t *testing.T) {
	assert := assert.New(t)
	fs := afero.Afero{
		Fs: afero.NewMemMapFs(),
	}
	reader := KubeconfigReader{fs}
	_, err := reader.ReadKubeconfig()

	assert.Error(err)
}
