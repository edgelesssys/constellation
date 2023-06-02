/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteVersion(t *testing.T) {
	client := &fakeAttestationClient{}
	sut := deleteCmd{
		attestationclient: client,
	}
	cmd := newDeleteCmd()
	require.NoError(t, cmd.PersistentFlags().Set("version", "2021-01-01"))
	assert.NoError(t, sut.delete(context.Background(), cmd))
	assert.True(t, client.isCalled)
}

type fakeAttestationClient struct {
	isCalled bool
}

func (f *fakeAttestationClient) DeleteAzureSEVSNVersion(_ context.Context, version string) error {
	if version == "2021-01-01" {
		f.isCalled = true
		return nil
	}
	return errors.New("version does not exist")
}

func (f fakeAttestationClient) UploadAzureSEVSNP(_ context.Context, _ attestationconfig.AzureSEVSNPVersion, _ time.Time) error {
	return nil
}

func (f fakeAttestationClient) DeleteList(_ context.Context, _ variant.Variant) error {
	return nil
}

func (f fakeAttestationClient) List(_ context.Context, _ variant.Variant) ([]string, error) {
	return []string{}, nil
}
