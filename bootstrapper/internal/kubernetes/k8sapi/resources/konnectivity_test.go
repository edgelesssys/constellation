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

func TestKonnectivityMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	kmsDepl := NewKonnectivityAgents("192.168.2.1")
	data, err := kmsDepl.Marshal()
	require.NoError(err)

	var recreated konnectivityAgents
	require.NoError(kubernetes.UnmarshalK8SResources(data, &recreated))
	assert.Equal(kmsDepl, &recreated)
}
