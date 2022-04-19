package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/state"
)

type statusWaiter interface {
	InitializeValidators([]atls.Validator)
	WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error
}
