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

func TestAutoscalerDeploymentMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	autoscalerDepl := NewDefaultAutoscalerDeployment(nil, nil, nil, "")

	data, err := autoscalerDepl.Marshal()
	require.NoError(err)

	t.Log(string(data))

	var recreated AutoscalerDeployment
	require.NoError(kubernetes.UnmarshalK8SResources(data, &recreated))
	assert.Equal(autoscalerDepl, &recreated)
}

func TestAutoscalerDeploymentWithCommandMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	autoscalerDepl := NewDefaultAutoscalerDeployment(nil, nil, nil, "")

	data, err := autoscalerDepl.Marshal()
	require.NoError(err)

	t.Log(string(data))

	var recreated AutoscalerDeployment
	require.NoError(kubernetes.UnmarshalK8SResources(data, &recreated))
	assert.Equal(autoscalerDepl, &recreated)
}
