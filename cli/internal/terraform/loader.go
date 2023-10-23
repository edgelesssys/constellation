/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"embed"
	"errors"
	"io/fs"
	slashpath "path"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

//go:embed terraform/*
//go:embed terraform/*/.terraform.lock.hcl
//go:embed terraform/iam/*/.terraform.lock.hcl
var terraformFS embed.FS

// prepareWorkspace loads the embedded Terraform files,
// and writes them into the workspace.
func prepareWorkspace(rootDir string, fileHandler file.Handler, workingDir string) error {
	return terraformCopier(fileHandler, rootDir, workingDir)
}

// terraformCopier copies the embedded Terraform files into the workspace.
// allowOverwrites allows overwriting existing files in the workspace.
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
		opts := []file.Option{
			file.OptMkdirAll,
			// Allow overwriting existing files.
			// If we are creating a new cluster, the workspace must have been empty before,
			// so there is no risk of overwriting existing files.
			// If we are upgrading an existing cluster, we want to overwrite the existing files,
			// and we have already created a backup of the existing workspace.
			file.OptOverwrite,
		}
		return fileHandler.Write(fileName, content, opts...)
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
