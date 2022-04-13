package cmd

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/coordinator/state"
)

type stubStatusWaiter struct {
	initialized   bool
	waitForAllErr error
}

func (s *stubStatusWaiter) InitializePCRs(gcpPCRs, azurePCRs map[uint32][]byte) {
	s.initialized = true
}

func (s *stubStatusWaiter) WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error {
	if !s.initialized {
		return errors.New("waiter not initialized")
	}
	return s.waitForAllErr
}
