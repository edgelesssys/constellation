package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/state"
)

type stubStatusWaiter struct {
	waitForAllErr error
}

func (w stubStatusWaiter) WaitForAll(ctx context.Context, status state.State, endpoints []string) error {
	return w.waitForAllErr
}
