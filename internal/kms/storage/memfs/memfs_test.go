/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package memfs

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

func TestMemMapStorage(t *testing.T) {
	assert := assert.New(t)

	store := New()

	testDEK1 := []byte("test DEK")
	testDEK2 := []byte("more test DEK")
	ctx := context.Background()

	// request unset value
	_, err := store.Get(ctx, "test:input")
	assert.ErrorIs(err, storage.ErrDEKUnset)

	// test Put method
	assert.NoError(store.Put(ctx, "volume01", testDEK1))
	assert.NoError(store.Put(ctx, "volume02", testDEK2))

	// make sure values have been set
	val, err := store.Get(ctx, "volume01")
	assert.NoError(err)
	assert.Equal(testDEK1, val)
	val, err = store.Get(ctx, "volume02")
	assert.NoError(err)
	assert.Equal(testDEK2, val)

	_, err = store.Get(ctx, "invalid:key")
	assert.ErrorIs(err, storage.ErrDEKUnset)
}
