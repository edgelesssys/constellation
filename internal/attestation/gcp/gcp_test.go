//go:build gcp

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"encoding/json"
	"testing"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAttestation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	PCR0 := []byte{0x0f, 0x35, 0xc2, 0x14, 0x60, 0x8d, 0x93, 0xc7, 0xa6, 0xe6, 0x8a, 0xe7, 0x35, 0x9b, 0x4a, 0x8b, 0xe5, 0xa0, 0xe9, 0x9e, 0xea, 0x91, 0x07, 0xec, 0xe4, 0x27, 0xc4, 0xde, 0xa4, 0xe4, 0x39, 0xcf}

	issuer := NewIssuer()
	validator := NewValidator(map[uint32][]byte{0: PCR0}, nil, nil)

	nonce := []byte{2, 3, 4}
	challenge := []byte("Constellation")

	attDocRaw, err := issuer.Issue(challenge, nonce)
	assert.NoError(err)

	var attDoc vtpm.AttestationDocument
	err = json.Unmarshal(attDocRaw, &attDoc)
	require.NoError(err)
	assert.Equal(challenge, attDoc.UserData)
	originalPCR := attDoc.Attestation.Quotes[1].Pcrs.Pcrs[uint32(vtpm.PCRIndexOwnerID)]

	out, err := validator.Validate(attDocRaw, nonce)
	assert.NoError(err)
	assert.Equal(challenge, out)

	// Mark node as intialized. We should still be abe to validate
	assert.NoError(vtpm.MarkNodeAsBootstrapped(vtpm.OpenVTPM, []byte("Test")))

	attDocRaw, err = issuer.Issue(challenge, nonce)
	assert.NoError(err)

	// Make sure the PCR changed
	err = json.Unmarshal(attDocRaw, &attDoc)
	require.NoError(err)
	assert.NotEqual(originalPCR, attDoc.Attestation.Quotes[1].Pcrs.Pcrs[uint32(vtpm.PCRIndexOwnerID)])

	out, err = validator.Validate(attDocRaw, nonce)
	assert.NoError(err)
	assert.Equal(challenge, out)
}
