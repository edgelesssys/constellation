package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoscalingNodeGroup(t *testing.T) {
	assert := assert.New(t)
	nodeGroups := AutoscalingNodeGroup("some-project", "some-zone", "some-group", 0, 100)
	wantNodeGroups := "0:100:https://www.googleapis.com/compute/v1/projects/some-project/zones/some-zone/instanceGroups/some-group"
	assert.Equal(wantNodeGroups, nodeGroups)
}
