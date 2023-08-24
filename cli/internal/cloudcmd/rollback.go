/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// rollbacker does a rollback.
type rollbacker interface {
	rollback(ctx context.Context, w io.Writer, logLevel terraform.LogLevel) error
}

// rollbackOnError calls rollback on the rollbacker if the handed error is not nil,
// and writes logs to the writer w.
func rollbackOnError(w io.Writer, onErr *error, roll rollbacker, logLevel terraform.LogLevel) {
	if *onErr == nil {
		return
	}
	fmt.Fprintf(w, "An error occurred: %s\n", *onErr)
	fmt.Fprintln(w, "Attempting to roll back.")
	if err := roll.rollback(context.Background(), w, logLevel); err != nil {
		*onErr = errors.Join(*onErr, fmt.Errorf("on rollback: %w", err)) // TODO(katexochen): print the error, or return it?
		return
	}
	fmt.Fprintln(w, "Rollback succeeded.")
}

type rollbackerTerraform struct {
	client tfCommonClient
}

func (r *rollbackerTerraform) rollback(ctx context.Context, w io.Writer, logLevel terraform.LogLevel) error {
	if err := r.client.Destroy(ctx, logLevel); err != nil {
		fmt.Fprintf(w, "Could not destroy the resources. Please delete the %q directory manually if no resources were created\n",
			constants.TerraformWorkingDir)
		return err
	}
	return r.client.CleanUpWorkspace()
}

type rollbackerQEMU struct {
	client           tfResourceClient
	libvirt          libvirtRunner
	createdWorkspace bool
}

func (r *rollbackerQEMU) rollback(ctx context.Context, w io.Writer, logLevel terraform.LogLevel) (retErr error) {
	if r.createdWorkspace {
		retErr = r.client.Destroy(ctx, logLevel)
	}
	if retErr := errors.Join(retErr, r.libvirt.Stop(ctx)); retErr != nil {
		fmt.Fprintf(w, "Could not destroy the resources. Please delete the %q directory manually if no resources were created\n",
			constants.TerraformWorkingDir)
		return retErr
	}
	return r.client.CleanUpWorkspace()
}
