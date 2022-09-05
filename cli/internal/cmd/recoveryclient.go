/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"io"

	"github.com/edgelesssys/constellation/internal/atls"
)

type recoveryClient interface {
	Connect(endpoint string, validators atls.Validator) error
	PushStateDiskKey(ctx context.Context, stateDiskKey, measurementSecret []byte) error
	io.Closer
}
