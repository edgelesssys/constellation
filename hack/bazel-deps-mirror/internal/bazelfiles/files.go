/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package bazelfiles is used to find and handle Bazel WORKSPACE and bzl files.
package bazelfiles

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bazelbuild/buildtools/build"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/spf13/afero"
)

// Helper is used to find and handle Bazel WORKSPACE and bzl files.
type Helper struct {
	fs            afero.Fs
	workspaceRoot string
}

// New creates a new BazelFilesHelper.
func New() (*Helper, error) {
	workspaceRoot, err := findWorkspaceRoot(os.LookupEnv)
	if err != nil {
		return nil, err
	}

	return &Helper{
		fs:            afero.NewBasePathFs(afero.NewOsFs(), workspaceRoot),
		workspaceRoot: workspaceRoot,
	}, nil
}

// FindFiles returns the paths to all Bazel files in the Bazel workspace.
func (h *Helper) FindFiles() ([]BazelFile, error) {
	workspaceFile, err := h.findWorkspaceFile()
	if err != nil {
		return nil, err
	}

	bzlFiles, err := h.findBzlFiles()
	if err != nil {
		return nil, err
	}

	return append(bzlFiles, workspaceFile), nil
}

// findWorkspaceFile returns the path to the Bazel WORKSPACE.bzlmod file (or WORKSPACE if the former doesn't exist).
func (h *Helper) findWorkspaceFile() (BazelFile, error) {
	if _, err := h.fs.Stat("WORKSPACE.bzlmod"); err == nil {
		return BazelFile{
			RelPath: "WORKSPACE.bzlmod",
			AbsPath: filepath.Join(h.workspaceRoot, "WORKSPACE.bzlmod"),
			Type:    BazelFileTypeWorkspace,
		}, nil
	}
	if _, err := h.fs.Stat("WORKSPACE"); err == nil {
		return BazelFile{
			RelPath: "WORKSPACE",
			AbsPath: filepath.Join(h.workspaceRoot, "WORKSPACE"),
			Type:    BazelFileTypeWorkspace,
		}, nil
	}
	return BazelFile{}, fmt.Errorf("failed to find Bazel WORKSPACE file")
}

// findBzlFiles returns the paths to all .bzl files in the Bazel workspace.
func (h *Helper) findBzlFiles() ([]BazelFile, error) {
	var bzlFiles []BazelFile
	err := afero.Walk(h.fs, ".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".bzl" {
			return nil
		}
		bzlFiles = append(bzlFiles, BazelFile{
			RelPath: path,
			AbsPath: filepath.Join(h.workspaceRoot, path),
			Type:    BazelFileTypeBzl,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bzlFiles, nil
}

// LoadFile loads a Bazel file.
func (h *Helper) LoadFile(bf BazelFile) (*build.File, error) {
	data, err := afero.ReadFile(h.fs, bf.RelPath)
	if err != nil {
		return nil, err
	}
	switch bf.Type {
	case BazelFileTypeBzl:
		return build.ParseBzl(bf.AbsPath, data)
	case BazelFileTypeWorkspace:
		return build.ParseWorkspace(bf.AbsPath, data)
	}
	return nil, fmt.Errorf("unknown Bazel file type: %d", bf.Type)
}

// WriteFile writes (updates) a Bazel file.
func (h *Helper) WriteFile(bf BazelFile, buildfile *build.File) error {
	return afero.WriteFile(h.fs, bf.RelPath, build.Format(buildfile), 0o644)
}

// Diff returns the diff between the saved and the updated (in-memory) version of a Bazel file.
func (h *Helper) Diff(bf BazelFile, buildfile *build.File) (string, error) {
	savedData, err := afero.ReadFile(h.fs, bf.RelPath)
	if err != nil {
		return "", err
	}
	updatedData := build.Format(buildfile)
	edits := myers.ComputeEdits(span.URIFromPath(bf.RelPath), string(savedData), string(updatedData))
	diff := fmt.Sprint(gotextdiff.ToUnified("a/"+bf.RelPath, "b/"+bf.RelPath, string(savedData), edits))
	return diff, nil
}

// findWorkspaceRoot returns the path to the Bazel workspace root.
func findWorkspaceRoot(lookupEnv LookupEnv) (string, error) {
	workspaceRoot, ok := lookupEnv("BUILD_WORKSPACE_DIRECTORY")
	if !ok {
		return "", fmt.Errorf("failed to find Bazel workspace root: not executed via \"bazel run\" and BUILD_WORKSPACE_DIRECTORY not set")
	}
	return workspaceRoot, nil
}

// BazelFile is a reference (path) to a Bazel file.
type BazelFile struct {
	RelPath string
	AbsPath string
	Type    BazelFileType
}

// BazelFileType is the type of a Bazel file.
type BazelFileType int

const (
	// BazelFileTypeBzl is a .bzl file.
	BazelFileTypeBzl = iota
	// BazelFileTypeWorkspace is a WORKSPACE or WORKSPACE.bzlmod file.
	BazelFileTypeWorkspace
)

// LookupEnv can be the real os.LookupEnv or a mock for testing.
type LookupEnv func(key string) (string, bool)
