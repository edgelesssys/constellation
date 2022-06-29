package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerificationDaemonset(t *testing.T) {
	deployment := NewVerificationDaemonSet("csp")
	deploymentYAML, err := deployment.Marshal()
	require.NoError(t, err)

	var recreated verificationDaemonset
	require.NoError(t, UnmarshalK8SResources(deploymentYAML, &recreated))
	assert.Equal(t, deployment, &recreated)
}
