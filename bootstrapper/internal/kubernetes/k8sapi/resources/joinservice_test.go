package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJoinServiceDaemonset(t *testing.T) {
	deployment := NewJoinServiceDaemonset("csp", "measurementsJSON", "idJSON")
	deploymentYAML, err := deployment.Marshal()
	require.NoError(t, err)

	var recreated joinServiceDaemonset
	require.NoError(t, UnmarshalK8SResources(deploymentYAML, &recreated))
	assert.Equal(t, deployment, &recreated)
}
