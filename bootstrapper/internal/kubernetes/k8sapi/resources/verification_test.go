/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerificationDaemonset(t *testing.T) {
	deployment := NewVerificationDaemonSet("csp", "192.168.2.1")
	deploymentYAML, err := deployment.Marshal()
	require.NoError(t, err)

	var recreated verificationDaemonset
	require.NoError(t, kubernetes.UnmarshalK8SResources(deploymentYAML, &recreated))
	assert.Equal(t, deployment, &recreated)
}
