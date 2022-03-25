package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s "k8s.io/api/core/v1"
)

func TestCloudControllerMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	cloudControllerManagerDepl := NewDefaultCloudControllerManagerDeployment("dummy-cloudprovider", "some-image:latest", "/dummy_path", []string{}, []k8s.Volume{}, []k8s.VolumeMount{}, nil)
	data, err := cloudControllerManagerDepl.Marshal()
	require.NoError(err)

	var recreated cloudControllerManagerDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(cloudControllerManagerDepl, &recreated)
}
