package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlannelDeployment(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	flannelDeployment := NewDefaultFlannelDeployment()
	data, err := flannelDeployment.Marshal()
	require.NoError(err)

	var recreated FlannelDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(flannelDeployment, &recreated)
}
