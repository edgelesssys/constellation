package fallback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDiscoverDebugdIPs(t *testing.T) {
	assert := assert.New(t)

	fetcher := Fetcher{}
	ips, err := fetcher.DiscoverDebugdIPs(context.Background())

	assert.NoError(err)
	assert.Empty(ips)
}

func TestFetchSSHKeys(t *testing.T) {
	assert := assert.New(t)

	fetcher := Fetcher{}
	keys, err := fetcher.FetchSSHKeys(context.Background())

	assert.NoError(err)
	assert.Empty(keys)
}
