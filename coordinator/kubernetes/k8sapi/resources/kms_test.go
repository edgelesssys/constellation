package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKMSMarshalUnmarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	testMS := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}
	kmsDepl := NewKMSDeployment(testMS)
	data, err := kmsDepl.Marshal()
	require.NoError(err)

	var recreated kmsDeployment
	require.NoError(UnmarshalK8SResources(data, &recreated))
	assert.Equal(kmsDepl, &recreated)
}
