/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cmd provides the Constellation CLI.

It is responsible for the interaction with the user.
*/
package cmd

import (
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

func configPath(workspace string) string {
	return filepath.Join(workspace, constants.ConfigFilename)
}
