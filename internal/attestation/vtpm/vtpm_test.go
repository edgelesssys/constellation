package vtpm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNOPTPM(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(MarkNodeAsInitialized(OpenNOPTPM, []byte{0x0, 0x1, 0x2, 0x3}, []byte{0x4, 0x5, 0x6, 0x7}))
}
