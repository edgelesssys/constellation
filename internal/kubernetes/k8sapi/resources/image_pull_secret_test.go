package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImagePullSecret(t *testing.T) {
	imgPullSec := NewImagePullSecret("namespace")
	_, err := imgPullSec.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, "namespace", imgPullSec.Namespace)
}
