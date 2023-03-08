/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package journald provides functions to read and collect journald logs.
*/
package journald

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type command interface {
	Output() ([]byte, error)
}

// Collector represents a command that is executed by journalctl.
type Collector struct {
	cmd command
}

// NewCollector creates a new JournaldCommand.
func NewCollector(ctx context.Context, service string) (*Collector, error) {
	cmd := exec.CommandContext(ctx, "journalctl", "-u", service)
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	return &Collector{cmd}, nil
}

// Collect gets all journald logs from a service and returns a byte array with the plain text logs.
func (c *Collector) Collect() ([]byte, error) {
	out, err := c.cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("executing %q failed: %s", c.cmd, exitErr.Stderr)
	} else if err != nil {
		return nil, fmt.Errorf("executing command: %w", err)
	}
	return out, nil
}
