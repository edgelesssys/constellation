/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package fallback

import (
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
	ips, err := fetcher.DiscoverDebugdIPs(t.Context())
	assert.NoError(err)
	assert.Empty(ips)

	rol, err := fetcher.Role(t.Context())
	assert.NoError(err)
	assert.Equal(rol, role.Unknown)

	uid, err := fetcher.UID(t.Context())
	assert.NoError(err)
	assert.Empty(uid)

	self, err := fetcher.Self(t.Context())
	assert.NoError(err)
	assert.Empty(self)
}
