/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAttestationKey(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()

	// create the attestation key in RSA format
	tpmAk, err := tpmclient.AttestationKeyRSA(tpm)
	assert.NoError(err)
	assert.NotNil(tpmAk)

	// get the cached, already created key
	getAk, err := getAttestationKey(tpm)
	assert.NoError(err)
	assert.NotNil(getAk)

	// if everything worked fine, tpmAk and getAk are the same key
	assert.Equal(tpmAk, getAk)
}
