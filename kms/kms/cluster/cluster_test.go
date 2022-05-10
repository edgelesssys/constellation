package cluster

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordinatorKMS(t *testing.T) {
	assert := assert.New(t)
	kms := &ClusterKMS{}
	masterKey := []byte("Constellation")

	key, err := kms.GetDEK(context.Background(), "", "key-1", 32)
	assert.Error(err)
	assert.Nil(key)

	err = kms.CreateKEK(context.Background(), "", masterKey)
	assert.NoError(err)
	assert.Equal(masterKey, kms.masterKey)

	key1, err := kms.GetDEK(context.Background(), "", "key-1", 32)
	assert.NoError(err)
	assert.Len(key1, 32)

	key2, err := kms.GetDEK(context.Background(), "", "key-1", 32)
	assert.NoError(err)
	assert.Equal(key1, key2)

	key3, err := kms.GetDEK(context.Background(), "", "key-2", 32)
	assert.NoError(err)
	assert.NotEqual(key1, key3)

	key, err = kms.GetDEK(context.Background(), "", "key", 64)
	assert.NoError(err)
	assert.Len(key, 64)
}
