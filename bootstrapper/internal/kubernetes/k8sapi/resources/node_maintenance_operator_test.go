package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeMaintenanceOperatorMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	nmoDepl := NewNodeMaintenanceOperatorDeployment()
	data, err := nmoDepl.Marshal()
	require.NoError(err)

	var recreated nodeMaintenanceOperatorDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(nmoDepl, &recreated)
}
