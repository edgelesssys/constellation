/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package nodelock

import (
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/stretchr/testify/assert"
)

func TestTryLockOnce(t *testing.T) {
	assert := assert.New(t)
	tpm := spyDevice{}
	lock := Lock{
		tpm:    tpm.Opener(),
		marker: stubMarker,
	}
	locked, err := lock.TryLockOnce(nil)
	assert.NoError(err)
	assert.True(locked)

	wg := sync.WaitGroup{}
	tryLock := func() {
		defer wg.Done()
		locked, err := lock.TryLockOnce(nil)
		assert.NoError(err)
		assert.False(locked)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go tryLock()
	}

	wg.Wait()

	assert.EqualValues(1, tpm.counter.Load())
}

type spyDevice struct {
	counter atomic.Uint64
}

func (s *spyDevice) Opener() vtpm.TPMOpenFunc {
	return func() (io.ReadWriteCloser, error) {
		s.counter.Add(1)
		return nil, nil
	}
}

func stubMarker(openDevice func() (io.ReadWriteCloser, error), _ []byte) error {
	// this only needs to invoke the openDevice function
	// so that the spyTPM counter is incremented
	_, _ = openDevice()
	return nil
}
