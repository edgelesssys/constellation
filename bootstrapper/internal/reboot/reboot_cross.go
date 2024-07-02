//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package reboot

// Reboot is not implemented on non-Linux platforms.
func Reboot(_ error) {
	panic("reboot not implemented on non-Linux platforms")
}
