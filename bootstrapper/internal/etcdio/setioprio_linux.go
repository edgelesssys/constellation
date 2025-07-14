//go:build linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package etcdio

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func setioprio(ioPrioWhoProcess, pid, prioVal uintptr) (uintptr, uintptr, syscall.Errno) {
	return unix.Syscall(unix.SYS_IOPRIO_SET, ioPrioWhoProcess, pid, prioVal)
}
