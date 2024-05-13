//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"log/slog"
)

func (c *Cloud) PrepareControlPlaneNode(_ context.Context, _ *slog.Logger) error {
	panic("azure.*Cloud.PrepareControlPlaneNode is only supported on Linux")
}
