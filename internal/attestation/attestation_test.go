/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestation

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/crypto/testvector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveClusterID(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	testvector := testvector.HKDFClusterID
	clusterID, err := DeriveClusterID(testvector.Secret, testvector.Salt)
	require.NoError(err)
	assert.Equal(testvector.Output, clusterID)

	clusterIDdiff, err := DeriveClusterID(testvector.Secret, []byte("different-salt"))
	require.NoError(err)
	assert.NotEqual(clusterID, clusterIDdiff)

	clusterIDdiff, err = DeriveClusterID([]byte("different-secret"), testvector.Salt)
	require.NoError(err)
	assert.NotEqual(clusterID, clusterIDdiff)
}
