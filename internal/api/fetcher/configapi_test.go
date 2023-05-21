package fetcher_test

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	ctx := context.Background()
	fetcher := fetcher.NewConfigAPIFetcher()
	res, err := fetcher.FetchLatestAzureSEVSNPVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint8(2), res.Bootloader)
}
