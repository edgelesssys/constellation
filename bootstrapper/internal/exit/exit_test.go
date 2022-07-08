package exit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	cleaner := New(&spyStopper{})
	assert.NotNil(cleaner)
	assert.NotEmpty(cleaner.stoppers)
}

func TestWith(t *testing.T) {
	assert := assert.New(t)

	cleaner := New().With(&spyStopper{})
	assert.NotEmpty(cleaner.stoppers)
}

func TestClean(t *testing.T) {
	assert := assert.New(t)

	stopper := &spyStopper{}
	cleaner := New(stopper)
	cleaner.Clean()
	assert.True(stopper.stopped)

	// call again to make sure it doesn't panic or block
	cleaner.Clean()
}

type spyStopper struct {
	stopped bool
}

func (s *spyStopper) Stop() {
	s.stopped = true
}
