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
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAccessManagerMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// Without data
	accessManagerDeplNil := NewAccessManagerDeployment(nil)
	data, err := accessManagerDeplNil.Marshal()
	require.NoError(err)

	var recreated AccessManagerDeployment
	require.NoError(kubernetes.UnmarshalK8SResources(data, &recreated))
	assert.Equal(accessManagerDeplNil, &recreated)

	// With data
	sshUsers := make(map[string]string)
	sshUsers["test-user"] = "ssh-rsa abcdefg"
	accessManagerDeplNil = NewAccessManagerDeployment(sshUsers)
	data, err = accessManagerDeplNil.Marshal()
	require.NoError(err)

	require.NoError(kubernetes.UnmarshalK8SResources(data, &recreated))
	assert.Equal(accessManagerDeplNil, &recreated)
}
