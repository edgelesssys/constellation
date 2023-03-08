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
	"os/exec"
)

type collectionCommand interface {
	executeCommand() ([]byte, error)
}

// Collector represents a command that is executed by journalctl.
type Collector struct {
	command *exec.Cmd
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
func Collect(jcmd collectionCommand) ([]byte, error) {
	out, err := jcmd.executeCommand()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (j *Collector) executeCommand() ([]byte, error) {
	return j.command.Output()
}
