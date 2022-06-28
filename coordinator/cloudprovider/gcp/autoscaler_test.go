package gcp

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/stretchr/testify/assert"
)

func TestTrivialAutoscalerFunctions(t *testing.T) {
	assert := assert.New(t)
	autoscaler := Autoscaler{}

	assert.NotEmpty(autoscaler.Name())
	assert.Empty(autoscaler.Secrets(metadata.InstanceMetadata{}, ""))
	assert.NotEmpty(autoscaler.Volumes())
	assert.NotEmpty(autoscaler.VolumeMounts())
	assert.NotEmpty(autoscaler.Env())
	assert.True(autoscaler.Supported())
}
