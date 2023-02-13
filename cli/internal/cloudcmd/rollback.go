/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

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
		*onErr = multierr.Append(*onErr, fmt.Errorf("on rollback: %w", err)) // TODO: print the error, or return it?
		return
	}
	fmt.Fprintln(w, "Rollback succeeded.")
}

type rollbackerTerraform struct {
	client terraformClient
}

func (r *rollbackerTerraform) rollback(ctx context.Context) error {
	var err error
	err = multierr.Append(err, r.client.Destroy(ctx))
	if err == nil {
		err = multierr.Append(err, r.client.CleanUpWorkspace())
	}
	return err
}

type rollbackerQEMU struct {
	client           terraformClient
	libvirt          libvirtRunner
	createdWorkspace bool
}

func (r *rollbackerQEMU) rollback(ctx context.Context) error {
	var err error
	if r.createdWorkspace {
		err = multierr.Append(err, r.client.Destroy(ctx))
	}
	err = multierr.Append(err, r.libvirt.Stop(ctx))
	if err == nil {
		err = r.client.CleanUpWorkspace()
	}
	return err
}
