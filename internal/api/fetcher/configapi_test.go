/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package fetcher_test

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	cancel := configapi.UseDummyConfigAPIServer(8081)
	defer cancel()
	fetcher := fetcher.NewConfigAPIFetcher()
	res, err := fetcher.FetchLatestAzureSEVSNPVersion(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(2), res.Bootloader)
}
