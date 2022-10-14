/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"testing"

	"github.com/google/go-tpm-tools/simulator"
	"github.com/stretchr/testify/require"
)

func TestGetAWSInstanceInfo(t *testing.T) {
	t.Skip("aws validator not implemented")

}

func TestGetAttestationKey(t *testing.T) {
	require := require.New(t)
	//assert := assert.New(t)

	tpm, err := simulator.Get()
	require.NoError(err)
	defer tpm.Close()
}

type fakeMetadataClient struct {
	projectIDString    string
	instanceNameString string
	zoneString         string
	projecIDErr        error
	instanceNameErr    error
	zoneErr            error
}
