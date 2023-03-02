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

// Command represents a command that is executed by journalctl.
type Command struct {
	ctx     context.Context
	service string
	command *exec.Cmd
}

// NewCommand creates a new JournaldCommand.
func NewCommand(ctx context.Context, service string) (*Command, error) {
	cmd := exec.CommandContext(ctx, "journalctl", "-u", service)
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	return &Command{ctx, service, cmd}, nil
}

// GetServiceLog gets all journald logs from a service and returns a byte array with the plain text logs.
func GetServiceLog(jcmd collectionCommand) ([]byte, error) {
	out, err := jcmd.executeCommand()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (j *Command) executeCommand() ([]byte, error) {
	return j.command.Output()
}
