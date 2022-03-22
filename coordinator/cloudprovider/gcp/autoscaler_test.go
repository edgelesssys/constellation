package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrivialAutoscalerFunctions(t *testing.T) {
	assert := assert.New(t)
	autoscaler := Autoscaler{}

	assert.NotEmpty(autoscaler.Name())
	assert.True(autoscaler.Supported())
}
