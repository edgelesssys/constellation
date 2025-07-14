//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package azure

import (
	"context"
	"log/slog"
)

// PrepareControlPlaneNode is only supported on Linux.
func (c *Cloud) PrepareControlPlaneNode(_ context.Context, _ *slog.Logger) error {
	panic("azure.*Cloud.PrepareControlPlaneNode is only supported on Linux")
}
