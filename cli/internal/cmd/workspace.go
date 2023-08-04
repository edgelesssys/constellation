/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// Users may override the default workspace using the --workspace flag.
// The default workspace is the current working directory.
// The following functions return paths relative to the set workspace,
// and should be used when printing the path to the user.
// The MUST not be used when accessing files, as the workspace is changed
// using os.Chdir() before the command is executed.

func AdminConfPath(workspace string) string {
	return filepath.Join(workspace, constants.AdminConfFilename)
}

func ConfigPath(workspace string) string {
	return filepath.Join(workspace, constants.ConfigFilename)
}

func ClusterIDsPath(workspace string) string {
	return filepath.Join(workspace, constants.ClusterIDsFilename)
}

func MasterSecretPath(workspace string) string {
	return filepath.Join(workspace, constants.MasterSecretFilename)
}

func TerraformClusterWorkspace(workspace string) string {
	return filepath.Join(workspace, constants.TerraformWorkingDir)
}

func TerraformIAMWorkspace(workspace string) string {
	return filepath.Join(workspace, constants.TerraformIAMWorkingDir)
}

func TerraformLogPath(workspace string) string {
	return filepath.Join(workspace, constants.TerraformLogFile)
}

const gcpServiceAccountKeyFile = "gcpServiceAccountKey.json"

func GcpServiceAccountKeyPath(workspace string) string {
	return filepath.Join(workspace, gcpServiceAccountKeyFile)
}
