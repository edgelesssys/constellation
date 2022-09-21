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
	k8s "k8s.io/api/core/v1"
)

func TestCloudControllerMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	cloudControllerManagerDepl := NewDefaultCloudControllerManagerDeployment("dummy-cloudprovider", "some-image:latest", "/dummy_path", "192.0.2.0/24", []string{}, []k8s.Volume{}, []k8s.VolumeMount{}, nil)
	data, err := cloudControllerManagerDepl.Marshal()
	require.NoError(err)

	var recreated cloudControllerManagerDeployment
	require.NoError(kubernetes.UnmarshalK8SResources(data, &recreated))
	assert.Equal(cloudControllerManagerDepl, &recreated)
}
