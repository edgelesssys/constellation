//go:build !linux || !amd64

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"io"
	"runtime"
)

// checkSystemRequirements checks if the system meets the requirements for running a MiniConstellation cluster.
// This will always fail on non-linux/amd64 platforms.
func (m *miniUpCmd) checkSystemRequirements(_ io.Writer) error {
	return fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s, a linux/amd64 platform is required", runtime.GOOS, runtime.GOARCH)
}
