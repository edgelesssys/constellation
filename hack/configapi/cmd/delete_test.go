/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteVersion(t *testing.T) {
	client := &fakeAttestationClient{}
	sut := deleteCmd{
		attestationClient: client,
	}
	cmd := newDeleteCmd()
	require.NoError(t, cmd.Flags().Set("version", "2021-01-01"))
	assert.NoError(t, sut.delete(cmd))
	assert.True(t, client.isCalled)
}

type fakeAttestationClient struct {
	isCalled bool
}

func (f *fakeAttestationClient) DeleteAzureSEVSNPVersion(_ context.Context, version string) error {
	if version == "2021-01-01" {
		f.isCalled = true
		return nil
	}
	return errors.New("version does not exist")
}
