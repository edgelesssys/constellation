//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package reboot

// Reboot is a no-op on non-Linux platforms.
func Reboot(_ error) {}
