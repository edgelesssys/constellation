package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemMapStorage(t *testing.T) {
	assert := assert.New(t)

	storage := NewMemMapStorage()

	testDEK1 := []byte("test DEK")
	testDEK2 := []byte("more test DEK")
	ctx := context.Background()

	// request unset value
	_, err := storage.Get(ctx, "test:input")
	assert.Error(err)

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
	assert.Error(err)
	assert.True(errors.Is(err, ErrDEKUnset))
}
