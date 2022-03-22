package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestState(t *testing.T) {
	assert := assert.New(t)

	var st State
	assert.Equal(Uninitialized, st)
	assert.Equal(Uninitialized, st.Get())
	assert.NoError(st.Require(Uninitialized))
	assert.Error(st.Require(AcceptingInit))

	st.Advance(AcceptingInit)
	assert.Equal(AcceptingInit, st)
	assert.Equal(AcceptingInit, st.Get())
	assert.Error(st.Require(Uninitialized))
	assert.NoError(st.Require(AcceptingInit))

	assert.Panics(func() { st.Advance(Uninitialized) })
}
