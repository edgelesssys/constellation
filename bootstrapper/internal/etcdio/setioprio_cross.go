//go:build !linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package etcdio

import (
	"syscall"
)

func setioprio(_, _, _ uintptr) (uintptr, uintptr, syscall.Errno) {
	panic("setioprio not implemented on non-Linux platforms")
}
