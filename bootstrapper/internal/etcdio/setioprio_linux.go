//go:build linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package etcdio

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func setioprio(ioPrioWhoProcess, pid, prioVal uintptr) (uintptr, uintptr, syscall.Errno) {
	return unix.Syscall(unix.SYS_IOPRIO_SET, ioPrioWhoProcess, pid, prioVal)
}
