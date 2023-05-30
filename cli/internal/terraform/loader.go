/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	slashpath "path"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

// ErrTerraformWorkspaceDifferentFiles is returned when a re-used existing Terraform workspace has different files than the ones to be extracted (e.g. due to a version mix-up or incomplete writes).
var ErrTerraformWorkspaceDifferentFiles = errors.New("creating cluster: trying to overwrite an existing Terraform file with a different version")

//go:embed terraform/*
//go:embed terraform/*/.terraform.lock.hcl
var terraformFS embed.FS

// prepareWorkspace loads the embedded Terraform files,
// and writes them into the workspace.
func prepareWorkspace(rootDir string, fileHandler file.Handler, workingDir string) error {
	return terraformCopier(fileHandler, rootDir, workingDir)
}

// prepareUpgradeWorkspace takes the Terraform state file from the old workspace and the
// embedded Terraform files and writes them into the new workspace.
func prepareUpgradeWorkspace(rootDir string, fileHandler file.Handler, oldWorkingDir, newWorkingDir string) error {
	// backup old workspace
	if err := fileHandler.CopyDir(
		oldWorkingDir,
		filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeBackupDir),
	); err != nil {
		return fmt.Errorf("backing up old workspace: %w", err)
	}

	// copy state file
	if err := fileHandler.CopyFile(
		filepath.Join(oldWorkingDir, "terraform.tfstate"),
		filepath.Join(newWorkingDir, "terraform.tfstate"),
		file.OptMkdirAll,
	); err != nil {
		return fmt.Errorf("copying state file: %w", err)
	}

	return terraformCopier(fileHandler, rootDir, newWorkingDir)
}

// terraformCopier copies the embedded Terraform files into the workspace.
func terraformCopier(fileHandler file.Handler, rootDir, workingDir string) error {
	goEmbedRootDir := filepath.ToSlash(rootDir)
	return fs.WalkDir(terraformFS, goEmbedRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		goEmbedPath := filepath.ToSlash(path)
		content, err := terraformFS.ReadFile(goEmbedPath)
		if err != nil {
			return err
		}
		// normalize
		fileName := strings.Replace(slashpath.Join(workingDir, path), goEmbedRootDir+"/", "", 1)
		if err := fileHandler.Write(fileName, content, file.OptMkdirAll); errors.Is(err, afero.ErrFileExists) {
			// If a file already exists, check if it is identical. If yes, continue and don't write anything to disk.
			// If no, don't overwrite it and instead throw an error. The affected file could be from a different version,
			// provider, corrupted or manually modified in general.
			existingFileContent, err := fileHandler.Read(fileName)
			if err != nil {
				return err
			}

			if !bytes.Equal(content, existingFileContent) {
				return ErrTerraformWorkspaceDifferentFiles
			}
			return nil
		} else if err != nil {
			return err
		}

		return nil
	})
}

func cleanUpWorkspace(fileHandler file.Handler, workingDir string) error {
	return ignoreFileNotFoundErr(fileHandler.RemoveAll(workingDir))
}

// ignoreFileNotFoundErr ignores the error if it is a file not found error.
func ignoreFileNotFoundErr(err error) error {
	if errors.Is(err, afero.ErrFileNotFound) {
		return nil
	}
	return err
}
