package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoscalerDeploymentMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	autoscalerDepl := NewDefaultAutoscalerDeployment(nil, nil, nil, "")

	data, err := autoscalerDepl.Marshal()
	require.NoError(err)

	t.Log(string(data))

	var recreated autoscalerDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(autoscalerDepl, &recreated)
}

func TestAutoscalerDeploymentWithCommandMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	autoscalerDepl := NewDefaultAutoscalerDeployment(nil, nil, nil, "")
	autoscalerDepl.SetAutoscalerCommand("someProvider", []string{"group1", "group2"})

	data, err := autoscalerDepl.Marshal()
	require.NoError(err)

	t.Log(string(data))

	var recreated autoscalerDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(autoscalerDepl, &recreated)
}
