package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestMemMapStorage(t *testing.T) {
	assert := assert.New(t)

	storage := NewMemMapStorage()

	testDEK1 := []byte("test DEK")
	testDEK2 := []byte("more test DEK")
	ctx := context.Background()

	// request unset value
	_, err := storage.Get(ctx, "test:input")
	assert.ErrorIs(err, ErrDEKUnset)

	// test Put method
	assert.NoError(storage.Put(ctx, "volume01", testDEK1))
	assert.NoError(storage.Put(ctx, "volume02", testDEK2))

	// make sure values have been set
	val, err := storage.Get(ctx, "volume01")
	assert.NoError(err)
	assert.Equal(testDEK1, val)
	val, err = storage.Get(ctx, "volume02")
	assert.NoError(err)
	assert.Equal(testDEK2, val)

	_, err = storage.Get(ctx, "invalid:key")
	assert.ErrorIs(err, ErrDEKUnset)
}
