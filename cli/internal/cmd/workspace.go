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

func adminConfPath(workspace string) string {
	return filepath.Join(workspace, constants.AdminConfFilename)
}

func configPath(workspace string) string {
	return filepath.Join(workspace, constants.ConfigFilename)
}

func clusterIDsPath(workspace string) string {
	return filepath.Join(workspace, constants.ClusterIDsFilename)
}

func masterSecretPath(workspace string) string {
	return filepath.Join(workspace, constants.MasterSecretFilename)
}

func terraformClusterWorkspace(workspace string) string {
	return filepath.Join(workspace, constants.TerraformWorkingDir)
}

func terraformIAMWorkspace(workspace string) string {
	return filepath.Join(workspace, constants.TerraformIAMWorkingDir)
}

func terraformLogPath(workspace string) string {
	return filepath.Join(workspace, constants.TerraformLogFile)
}

const gcpServiceAccountKeyFile = "gcpServiceAccountKey.json"

func gcpServiceAccountKeyPath(workspace string) string {
	return filepath.Join(workspace, gcpServiceAccountKeyFile)
}
