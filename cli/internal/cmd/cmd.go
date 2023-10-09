/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cmd provides the Constellation CLI.

It is responsible for the interaction with the user.

Logic should be kept to input/output parsing whenever possible.
Any more complex code should usually be implemented in one of the other CLI packages.

The code here should be kept as cloud provider agnostic as possible.
Any CSP specific tasks should be handled by the "cloudcmd" package.

All filepaths handled by the CLI code should originate from here.
Common filepaths are defined as constants in the global "/internal/constants" package.
To generate workspace correct filepaths for printing, use the functions from the "workspace" package.
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/spf13/pflag"
)

// rootFlags are flags defined on the root command.
// They are available to all subcommands.
type rootFlags struct {
	pathPrefixer pathprefix.PathPrefixer
	tfLog        string
	debug        bool
	force        bool
}

// parse flags into the rootFlags struct.
func (f *rootFlags) parse(flags *pflag.FlagSet) error {
	var errs error

	workspace, err := flags.GetString("workspace")
	if err != nil {
		errs = errors.Join(err, fmt.Errorf("getting 'workspace' flag: %w", err))
	}
	f.pathPrefixer = pathprefix.New(workspace)

	f.tfLog, err = flags.GetString("tf-log")
	if err != nil {
		errs = errors.Join(err, fmt.Errorf("getting 'tf-log' flag: %w", err))
	}

	f.debug, err = flags.GetBool("debug")
	if err != nil {
		errs = errors.Join(err, fmt.Errorf("getting 'debug' flag: %w", err))
	}

	f.force, err = flags.GetBool("force")
	if err != nil {
		errs = errors.Join(err, fmt.Errorf("getting 'force' flag: %w", err))
	}
	return errs
}
