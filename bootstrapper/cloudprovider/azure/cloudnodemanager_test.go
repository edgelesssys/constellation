package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrivialCNMFunctions(t *testing.T) {
	assert := assert.New(t)
	cloud := CloudNodeManager{}

	assert.NotEmpty(cloud.Image("1.23"))
	assert.NotEmpty(cloud.Path())
	assert.NotEmpty(cloud.ExtraArgs())
	assert.True(cloud.Supported())
}
