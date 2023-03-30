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

// Collector collects logs from journald.
type Collector struct {
	cmd    command
	stdout io.ReadCloser
	stderr io.ReadCloser
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

// Pipe returns a pipe to read the systemd logs. This should be read with a bufio Reader.
func (c *Collector) Pipe() (io.ReadCloser, error) {
	if err := c.cmd.Start(); err != nil {
		return nil, err
	}
	return c.stdout, nil
}

// Error returns the error message of the journalctl command.
// The first two parameters are what's written to stderr as
// well as the exit/io error, the third one checks if the function
// ran successfully.
func (c *Collector) Error() ([]byte, error) {
	stderr, err := io.ReadAll(c.stderr)
	if err != nil {
		return nil, err
	}
	exitCode := c.cmd.Wait()
	return stderr, exitCode
}
