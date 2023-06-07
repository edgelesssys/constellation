//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package setup

import (
	"errors"
)

// Mount performs a mount syscall.
func (m DiskMounter) Mount(_ string, _ string, _ string, _ uintptr, _ string) error {
	return errors.New("mount not implemented on this platform")
}

// Unmount performs an unmount syscall.
func (m DiskMounter) Unmount(_ string, _ int) error {
	return errors.New("mount not implemented on this platform")
}
