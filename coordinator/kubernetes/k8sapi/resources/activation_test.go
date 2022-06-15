package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActivationDaemonset(t *testing.T) {
	deployment := NewActivationDaemonset("csp", "measurementsJSON", "idJSON")
	deploymentYAML, err := deployment.Marshal()
	require.NoError(t, err)

	var recreated activationDaemonset
	require.NoError(t, UnmarshalK8SResources(deploymentYAML, &recreated))
	assert.Equal(t, deployment, &recreated)
}
