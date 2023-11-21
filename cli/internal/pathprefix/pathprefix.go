/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package pathprefix is used to print correct filepaths for a configured workspace.

The default workspace is the current working directory.
Users may override the default workspace using the --workspace flag.

The functions defined here should be used when printing any filepath to the user,
as they might otherwise be incorrect if the user has changed the workspace.

The prefixer MUST not be used when accessing files, as the workspace is changed
using os.Chdir() before the command is executed.
*/
package pathprefix

import (
	"path/filepath"
)

// PathPrefixer is used to prefix paths with the configured workspace for printing.
type PathPrefixer struct {
	workspace string
}

// New returns a new PathPrefixer.
func New(workspace string) PathPrefixer {
	return PathPrefixer{workspace: workspace}
}

// PrefixPrintablePath prefixes the given path with the configured workspace for printing.
// This function MUST not be used when accessing files.
// This function SHOULD be used when printing paths to the user.
func (p PathPrefixer) PrefixPrintablePath(path string) string {
	return filepath.Clean(filepath.Join(p.workspace, path))
}
