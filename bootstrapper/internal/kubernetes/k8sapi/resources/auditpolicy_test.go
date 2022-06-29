package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditPolicyMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	auditPolicy := NewDefaultAuditPolicy()
	data, err := auditPolicy.Marshal()
	require.NoError(err)

	var recreated AuditPolicy
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(auditPolicy, &recreated)
}
