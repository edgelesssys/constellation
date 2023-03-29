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

// Collector collects logs from journald.
type Collector struct {
	cmd command
}

// NewCollector creates a new Collector for journald logs.
func NewCollector(ctx context.Context) (*Collector, error) {
	cmd := exec.CommandContext(ctx, "journalctl")
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	return &Collector{cmd}, nil
}

// Collect gets all journald logs from a service and returns a byte slice with the plain text logs.
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
