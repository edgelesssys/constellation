package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudControllerMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	cloudControllerManagerDepl := NewDefaultCloudControllerManagerDeployment("dummy-cloudprovider", "some-image:latest", "/dummy_path")
	data, err := cloudControllerManagerDepl.Marshal()
	require.NoError(err)

	var recreated cloudControllerManagerDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(cloudControllerManagerDepl, &recreated)
}
