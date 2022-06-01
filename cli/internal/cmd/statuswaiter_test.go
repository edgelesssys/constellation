package cmd

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/internal/atls"
)

type stubStatusWaiter struct {
	initialized   bool
	initializeErr error
	waitForAllErr error
}

func (s *stubStatusWaiter) InitializeValidators([]atls.Validator) error {
	s.initialized = true
	return s.initializeErr
}

func (s *stubStatusWaiter) WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error {
	if !s.initialized {
		return errors.New("waiter not initialized")
	}
	return s.waitForAllErr
}
