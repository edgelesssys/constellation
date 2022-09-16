/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"embed"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

//go:embed all:terraform/*
var terraformFS embed.FS

// tfLoader handles loading and removal of Terraform files embedded in the CLI.
type tfLoader struct {
	file     file.Handler
	provider cloudprovider.Provider
}

func newLoader(file file.Handler, provider cloudprovider.Provider) *tfLoader {
	return &tfLoader{
		file:     file,
		provider: provider,
	}
}

// prepareWorkspace loads the embedded Terraform files,
// and writes them into the workspace.
func (t *tfLoader) prepareWorkspace() error {
	startDir := path.Join("terraform", strings.ToLower(t.provider.String()))
	return t.readEmbeddedFs(terraformFS, startDir, t.writeEmbeddedFile)
}

// cleanUpWorkspace removes files that were loaded into the workspace.
func (t *tfLoader) cleanUpWorkspace() error {
	startDir := path.Join("terraform", strings.ToLower(t.provider.String()))
	if err := t.readEmbeddedFs(terraformFS, startDir, t.removeEmbeddedFile); err != nil {
		return err
	}
	if err := ignoreFileNotFoundErr(t.file.Remove("modules/instance_group")); err != nil {
		return err
	}
	if err := ignoreFileNotFoundErr(t.file.Remove("modules")); err != nil {
		return err
	}

	return nil
}

// readEmbeddedFs reads a embed.Fs and calls the given function for each file.
func (t *tfLoader) readEmbeddedFs(fs embed.FS, dir string, action func(string) error) error {
	if dir == "" {
		dir = "."
	}
	embedDir, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, d := range embedDir {
		// use `path.Join` instead of `filepath.Join`
		// this avoids accidentally using Windows path separators in the embedded files
		path := path.Join(dir, d.Name())

		if d.IsDir() {
			if err := t.readEmbeddedFs(fs, path, action); err != nil {
				return err
			}
		} else {
			if err := action(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeEmbeddedFile writes the given file from the embedded filesystem to the workspace.
func (t *tfLoader) writeEmbeddedFile(fileName string) error {
	content, err := terraformFS.ReadFile(fileName)
	if err != nil {
		return err
	}
	fileName = strings.TrimPrefix(fileName, fmt.Sprintf("terraform/%s/", strings.ToLower(t.provider.String())))
	return t.file.Write(fileName, content, file.OptMkdirAll)
}

// removeEmbeddedFile removes the given file from the workspace.
func (t *tfLoader) removeEmbeddedFile(fileName string) error {
	fileName = strings.TrimPrefix(fileName, fmt.Sprintf("terraform/%s/", strings.ToLower(t.provider.String())))
	if err := ignoreFileNotFoundErr(t.file.Remove(fileName)); err != nil {
		return err
	}
	return nil
}

// ignoreFileNotFoundErr ignores the error if it is a file not found error.
func ignoreFileNotFoundErr(err error) error {
	if errors.Is(err, afero.ErrFileNotFound) {
		return nil
	}
	return err
}
