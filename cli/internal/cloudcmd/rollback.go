package cloudcmd

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/multierr"
)

// rollbacker does a rollback.
type rollbacker interface {
	rollback(ctx context.Context) error
}

// rollbackOnError calls rollback on the rollbacker if the handed error is not nil,
// and writes logs to the writer w.
func rollbackOnError(ctx context.Context, w io.Writer, onErr *error, roll rollbacker) {
	if *onErr == nil {
		return
	}
	fmt.Fprintf(w, "An error occurred: %s\n", *onErr)
	fmt.Fprintln(w, "Attempting to roll back.")
	if err := roll.rollback(ctx); err != nil {
		*onErr = multierr.Append(*onErr, fmt.Errorf("on rollback: %w", err)) // TODO: print the error, or retrun it?
		return
	}
	fmt.Fprintln(w, "Rollback succeeded.")
}

type rollbackerGCP struct {
	client gcpclient
}

func (r *rollbackerGCP) rollback(ctx context.Context) error {
	var err error
	err = multierr.Append(err, r.client.TerminateLoadBalancers(ctx))
	err = multierr.Append(err, r.client.TerminateInstances(ctx))
	err = multierr.Append(err, r.client.TerminateFirewall(ctx))
	err = multierr.Append(err, r.client.TerminateVPCs(ctx))
	return err
}

type rollbackerAzure struct {
	client azureclient
}

func (r *rollbackerAzure) rollback(ctx context.Context) error {
	return r.client.TerminateResourceGroupResources(ctx)
}
