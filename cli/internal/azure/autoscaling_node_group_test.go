package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAutoscalingNodeGroup(t *testing.T) {
	assert := assert.New(t)
	nodeGroups := AutoscalingNodeGroup("scale-set", 0, 100)
	wantNodeGroups := "0:100:scale-set"
	assert.Equal(wantNodeGroups, nodeGroups)
}
