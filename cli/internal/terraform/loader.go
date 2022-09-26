/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"embed"
	"errors"
	"io/fs"
	"path"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

//go:embed terraform/*
var terraformFS embed.FS

// prepareWorkspace loads the embedded Terraform files,
// and writes them into the workspace.
func prepareWorkspace(fileHandler file.Handler, provider cloudprovider.Provider) error {
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
		fileName := strings.TrimPrefix(path, rootDir+"/")
		return fileHandler.Write(fileName, content, file.OptMkdirAll)
	})
}

// cleanUpWorkspace removes files that were loaded into the workspace.
func cleanUpWorkspace(fileHandler file.Handler, provider cloudprovider.Provider) error {
	rootDir := path.Join("terraform", strings.ToLower(provider.String()))
	return fs.WalkDir(terraformFS, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fileName := strings.TrimPrefix(path, rootDir+"/")
		return ignoreFileNotFoundErr(fileHandler.RemoveAll(fileName))
	})
}

// ignoreFileNotFoundErr ignores the error if it is a file not found error.
func ignoreFileNotFoundErr(err error) error {
	if errors.Is(err, afero.ErrFileNotFound) {
		return nil
	}
	return err
}
