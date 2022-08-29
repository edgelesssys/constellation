package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrivialCNMFunctions(t *testing.T) {
	assert := assert.New(t)
	cloud := CloudNodeManager{}

	assert.Empty(cloud.Image(""))
	assert.Empty(cloud.Path())
	assert.Empty(cloud.ExtraArgs())
	assert.False(cloud.Supported())
}
