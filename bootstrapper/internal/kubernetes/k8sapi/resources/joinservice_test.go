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

func TestNewJoinServiceDaemonset(t *testing.T) {
	deployment := NewJoinServiceDaemonset("csp", "measurementsJSON", "enforcedPCRsJSON", "deadbeef", "true", []byte{0x0, 0x1, 0x2})
	deploymentYAML, err := deployment.Marshal()
	require.NoError(t, err)

	var recreated joinServiceDaemonset
	require.NoError(t, kubernetes.UnmarshalK8SResources(deploymentYAML, &recreated))
	assert.Equal(t, deployment, &recreated)
}
