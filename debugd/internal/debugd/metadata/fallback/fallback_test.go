/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fallback

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestDiscoverDebugdIPs(t *testing.T) {
	assert := assert.New(t)

	fetcher := NewFallbackFetcher()
	ips, err := fetcher.DiscoverDebugdIPs(context.Background())
	assert.NoError(err)
	assert.Empty(ips)

	rol, err := fetcher.Role(context.Background())
	assert.NoError(err)
	assert.Equal(rol, role.Unknown)

	uid, err := fetcher.UID(context.Background())
	assert.NoError(err)
	assert.Empty(uid)

	self, err := fetcher.Self(context.Background())
	assert.NoError(err)
	assert.Empty(self)
}
