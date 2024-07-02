//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package etcdio

import (
	"syscall"
)

func setioprio(_, _, _ uintptr) (uintptr, uintptr, syscall.Errno) {
	panic("setioprio not implemented on non-Linux platforms")
}
