package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/internal/atls"
)

type statusWaiter interface {
	InitializeValidators([]atls.Validator) error
	WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error
}
