//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
	// TODO: Holding off on writing integration tests for rekor.go until leak is
	// fixed. Even with ignore goleak still breaks
	// See: https://github.com/sigstore/rekor/issues/1094
	// goleak.IgnoreTopFunction("internal/poll.runtime_pollWait")
}

// func TestRekorSearchByHash(t *testing.T) {
// 	assert := assert.New(t)
// 	require := require.New(t)

// 	rekor, err := NewRekor()
// 	require.NoError(err)

// 	uuids, err := rekor.SearchByHash(context.Background(), "362f8ecba72f4326afaba7f6635b3e058888692841848e5514357315be9528474b23f5dcccb82b13")
// 	assert.NoError(err)
// 	assert.NotEmpty(uuids)
// }
