package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKMSMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	kmsDepl := NewKMSDeployment("test", KMSConfig{MasterSecret: []byte{0x0, 0x1, 0x2}, Salt: []byte{0x3, 0x4, 0x5}})
	data, err := kmsDepl.Marshal()
	require.NoError(err)

	var recreated kmsDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(kmsDepl, &recreated)
}
