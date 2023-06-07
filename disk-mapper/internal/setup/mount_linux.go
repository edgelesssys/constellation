//go:build linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package setup

import (
	"syscall"
)

// Mount performs a mount syscall.
func (m DiskMounter) Mount(source string, target string, fstype string, flags uintptr, data string) error {
	return syscall.Mount(source, target, fstype, flags, data)
}

// Unmount performs an unmount syscall.
func (m DiskMounter) Unmount(target string, flags int) error {
	return syscall.Unmount(target, flags)
}
