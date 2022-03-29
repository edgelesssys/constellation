package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/state"
)

type stubStatusWaiter struct {
	waitForAllErr error
}

func (w stubStatusWaiter) WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error {
	return w.waitForAllErr
}
