package vtpm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNOPTPM(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(MarkNodeAsBootstrapped(OpenNOPTPM, []byte{0x0, 0x1, 0x2, 0x3}))
}
