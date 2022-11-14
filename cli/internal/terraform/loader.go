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
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

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
		return fileHandler.Write(fileName, content, file.OptMkdirAll)
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
