package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/state"
)

type statusWaiter interface {
	WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error
}
