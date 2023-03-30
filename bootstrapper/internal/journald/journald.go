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
	Wait() error
}

// Collection is a trait that must be implemented
// by all log collectors.
type Collection interface {
	Collect() ([]byte, error)
}

// Collector collects logs from journald.
type Collector struct {
	cmd        command
	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser
}

// NewCollector creates a new Collector for journald logs.
func NewCollector(ctx context.Context) (*Collector, error) {
	cmd := exec.CommandContext(ctx, "journalctl")
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	collector := Collector{cmd, stdoutPipe, stderrPipe}
	return &collector, nil
}

// Start returns a pipe to read the systemd logs. This should be read with a bufio Reader.
func (c *Collector) Start() (io.ReadCloser, error) {
	if err := c.cmd.Start(); err != nil {
		return nil, err
	}
	return c.stdoutPipe, nil
}

// Error returns output to stderr as bytes as well
// as the exit code in form of an error.
func (c *Collector) Error() ([]byte, error) {
	stderr, err := io.ReadAll(c.stderrPipe)
	if err != nil {
		return nil, err
	}
	exitCode := c.cmd.Wait()
	return stderr, exitCode
}
