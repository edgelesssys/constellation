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
	"io"
	"os/exec"
)

type command interface {
	Start() error
}

// Collector collects logs from journald.
type Collector struct {
	cmd  command
	pipe io.ReadCloser
}

// NewCollector creates a new Collector for journald logs.
func NewCollector(ctx context.Context) (*Collector, error) {
	cmd := exec.CommandContext(ctx, "journalctl")
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &Collector{cmd, pipe}, nil
}

// Pipe returns a pipe to read the systemd logs. This should be read with a bufio Reader.
func (c *Collector) Pipe() (io.ReadCloser, error) {
	if err := c.cmd.Start(); err != nil {
		return nil, err
	}
	return c.pipe, nil
}
