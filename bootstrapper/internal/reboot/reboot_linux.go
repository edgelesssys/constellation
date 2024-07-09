//go:build linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package reboot

import (
	"log/syslog"
	"syscall"
	"time"
)

// Reboot writes an error message to the system log and reboots the system.
// We call this instead of os.Exit() since failures in the bootstrapper usually require a node reset.
func Reboot(e error) {
	syslogWriter, err := syslog.New(syslog.LOG_EMERG|syslog.LOG_KERN, "bootstrapper")
	if err != nil {
		_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	}
	_ = syslogWriter.Err(e.Error())
	_ = syslogWriter.Emerg("bootstrapper has encountered a non recoverable error. Rebooting...")
	time.Sleep(time.Minute) // sleep to allow the message to be written to syslog and seen by the user

	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
