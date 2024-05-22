/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// The etcdio package provides utilities to manage etcd I/O.
package etcdio

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"

	"golang.org/x/sys/unix"
)

var (
	// ErrNoEtcdProcess is returned when no etcd process is found on the node.
	ErrNoEtcdProcess = errors.New("no etcd process found on node")
	// ErrMultipleEtcdProcesses is returned when multiple etcd processes are found on the node.
	ErrMultipleEtcdProcesses = errors.New("multiple etcd processes found on node")
)

const (
	// Tells the syscall that a process' priority is going to be set.
	// See https://elixir.bootlin.com/linux/v6.9.1/source/include/uapi/linux/ioprio.h#L54.
	ioPrioWhoProcess = 1

	// See https://elixir.bootlin.com/linux/v6.9.1/source/include/uapi/linux/ioprio.h#L11.
	ioPrioClassShift = 13
	ioPrioNrClasses  = 8
	ioPrioClassMask  = ioPrioNrClasses - 1
	ioPrioPrioMask   = (1 << ioPrioClassShift) - 1

	targetClass = 1 // Realtime IO class for best scheduling prio
	targetPrio  = 0 // Highest priority within the class
)

// Client is a client for managing etcd I/O.
type Client struct {
	log *slog.Logger
}

// NewClient creates a new etcd I/O management client.
func NewClient(log *slog.Logger) *Client {
	return &Client{log: log}
}

// PrioritizeIO tries to find the etcd process on the node and prioritizes its I/O.
func (c *Client) PrioritizeIO() error {
	// find etcd process(es)
	pid, err := c.findEtcdProcess()
	if err != nil {
		return fmt.Errorf("finding etcd process: %w", err)
	}

	// Highest realtime priority value for the etcd process, see https://elixir.bootlin.com/linux/v6.9.1/source/include/uapi/linux/ioprio.h
	// for the calculation details.
	prioVal := ((targetClass & ioPrioClassMask) << ioPrioClassShift) | (targetPrio & ioPrioPrioMask)

	// see https://man7.org/linux/man-pages/man2/ioprio_set.2.html
	ret, _, errno := unix.Syscall(unix.SYS_IOPRIO_SET, ioPrioWhoProcess, uintptr(pid), uintptr(prioVal))
	if ret != 0 {
		return fmt.Errorf("setting I/O priority for etcd: %w", errno)
	}

	return nil
}

// findEtcdProcess tries to find the etcd process on the node.
func (c *Client) findEtcdProcess() (int, error) {
	procDir, err := os.Open("/proc")
	if err != nil {
		return 0, fmt.Errorf("opening /proc: %w", err)
	}
	defer procDir.Close()

	procEntries, err := procDir.Readdirnames(0)
	if err != nil {
		return 0, fmt.Errorf("reading /proc: %w", err)
	}

	// find etcd process(es)
	etcdPIDs := []int{}
	for _, f := range procEntries {
		// exclude non-pid dirs
		if f[0] < '0' || f[0] > '9' {
			continue
		}

		exe, err := os.Readlink(fmt.Sprintf("/proc/%s/exe", f))
		if err != nil {
			continue
		}

		if path.Base(exe) != "etcd" {
			continue
		}

		pid, err := strconv.Atoi(f)
		if err != nil {
			continue
		}

		// add the PID to the list of etcd PIDs
		etcdPIDs = append(etcdPIDs, pid)
	}

	if len(etcdPIDs) == 0 {
		return 0, ErrNoEtcdProcess
	}

	if len(etcdPIDs) > 1 {
		return 0, ErrMultipleEtcdProcesses
	}

	return etcdPIDs[0], nil
}
