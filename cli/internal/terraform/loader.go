/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
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
func prepareWorkspace(fileHandler file.Handler, provider cloudprovider.Provider, workingDir string) error {
	// use path.Join to ensure no forward slashes are used to read the embedded FS
	rootDir := path.Join("terraform", strings.ToLower(provider.String()))
	return fs.WalkDir(terraformFS, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := terraformFS.ReadFile(path)
		if err != nil {
			return err
		}
		fileName := strings.Replace(filepath.Join(workingDir, path), rootDir+"/", "", 1)
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
