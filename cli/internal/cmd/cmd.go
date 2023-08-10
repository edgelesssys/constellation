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
