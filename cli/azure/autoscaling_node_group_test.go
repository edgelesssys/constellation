package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoscalingNodeGroup(t *testing.T) {
	assert := assert.New(t)
	nodeGroups := AutoscalingNodeGroup("scale-set", 0, 100)
	expectedNodeGroups := "0:100:scale-set"
	assert.Equal(expectedNodeGroups, nodeGroups)
}
