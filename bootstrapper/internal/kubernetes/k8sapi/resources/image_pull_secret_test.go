package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImagePullSecret(t *testing.T) {
	imgPullSec := NewImagePullSecret()
	_, err := imgPullSec.Marshal()
	assert.NoError(t, err)
}
