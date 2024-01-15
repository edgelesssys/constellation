//go:build linux && amd64

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// checkSystemRequirements checks if the system meets the requirements for running a MiniConstellation cluster.
// We do so by verifying that the host:
// - arch/os is linux/amd64.
// - has access to /dev/kvm.
// - has at least 4 CPU cores.
// - has at least 4GB of memory.
// - has at least 20GB of free disk space.
func (m *miniUpCmd) checkSystemRequirements(out io.Writer) error {
	// check arch/os
	if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
		return fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s, a linux/amd64 platform is required", runtime.GOOS, runtime.GOARCH)
	}

	m.log.Debug("Checked arch and os")
	// check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); err != nil {
		return fmt.Errorf("unable to access KVM device: %w", err)
	}
	m.log.Debug("Checked that /dev/kvm exists")
	// check CPU cores
	if runtime.NumCPU() < 4 {
		return fmt.Errorf("insufficient CPU cores: %d, at least 4 cores are required by MiniConstellation", runtime.NumCPU())
	}
	if runtime.NumCPU() < 6 {
		fmt.Fprintf(out, "WARNING: Only %d CPU cores available. This may cause performance issues.\n", runtime.NumCPU())
	}
	m.log.Debug(fmt.Sprintf("Checked CPU cores - there are %d", runtime.NumCPU()))

	// check memory
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return fmt.Errorf("determining available memory: failed to open /proc/meminfo: %w", err)
	}
	defer f.Close()
	var memKB int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "MemTotal:") {
			_, err = fmt.Sscanf(scanner.Text(), "MemTotal:%d", &memKB)
			if err != nil {
				return fmt.Errorf("determining available memory: failed to parse /proc/meminfo: %w", err)
			}
		}
	}
	m.log.Debug("Scanned for available memory")
	memGB := memKB / 1024 / 1024
	if memGB < 4 {
		return fmt.Errorf("insufficient memory: %dGB, at least 4GB of memory are required by MiniConstellation", memGB)
	}
	if memGB < 6 {
		fmt.Fprintln(out, "WARNING: Less than 6GB of memory available. This may cause performance issues.")
	}
	m.log.Debug(fmt.Sprintf("Checked available memory, you have %dGB available", memGB))

	var stat unix.Statfs_t
	if err := unix.Statfs(".", &stat); err != nil {
		return err
	}
	freeSpaceGB := stat.Bavail * uint64(stat.Bsize) / 1024 / 1024 / 1024
	if freeSpaceGB < 20 {
		return fmt.Errorf("insufficient disk space: %dGB, at least 20GB of disk space are required by MiniConstellation", freeSpaceGB)
	}
	m.log.Debug(fmt.Sprintf("Checked for free space available, you have %dGB available", freeSpaceGB))

	return nil
}
