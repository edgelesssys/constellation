package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeOperatorMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	nmoDepl := NewNodeOperatorDeployment("csp", "uid")
	data, err := nmoDepl.Marshal()
	require.NoError(err)

	var recreated nodeOperatorDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(nmoDepl, &recreated)
}
