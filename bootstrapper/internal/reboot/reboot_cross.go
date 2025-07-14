//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package reboot

// Reboot is not implemented on non-Linux platforms.
func Reboot(_ error) {
	panic("reboot not implemented on non-Linux platforms")
}
