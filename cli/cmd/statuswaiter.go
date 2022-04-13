package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/state"
)

type statusWaiter interface {
	InitializePCRs(map[uint32][]byte, map[uint32][]byte)
	WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error
}
