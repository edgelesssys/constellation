package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudNodeManagerMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	cloudNodeManagerDepl := NewDefaultCloudNodeManagerDeployment("image", "path", []string{})
	data, err := cloudNodeManagerDepl.Marshal()
	require.NoError(err)

	var recreated cloudNodeManagerDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(cloudNodeManagerDepl, &recreated)
}
