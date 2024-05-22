/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package bazelfiles

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/edit"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestFindFiles(t *testing.T) {
	testCases := map[string]struct {
		files     []string
		wantFiles []BazelFile
		wantErr   bool
	}{
		"no WORKSPACE file": {
			files:     []string{},
			wantFiles: []BazelFile{},
			wantErr:   true,
		},
		"only WORKSPACE file": {
			files: []string{"WORKSPACE"},
			wantFiles: []BazelFile{
				{
					RelPath: "WORKSPACE",
					AbsPath: "/WORKSPACE",
					Type:    BazelFileTypeWorkspace,
				},
			},
		},
		"only WORKSPACE.bzlmod file": {
			files: []string{"WORKSPACE.bzlmod"},
			wantFiles: []BazelFile{
				{
					RelPath: "WORKSPACE.bzlmod",
					AbsPath: "/WORKSPACE.bzlmod",
					Type:    BazelFileTypeWorkspace,
				},
			},
		},
		"both WORKSPACE and WORKSPACE.bzlmod files": {
			files: []string{"WORKSPACE", "WORKSPACE.bzlmod"},
			wantFiles: []BazelFile{
				{
					RelPath: "WORKSPACE.bzlmod",
					AbsPath: "/WORKSPACE.bzlmod",
					Type:    BazelFileTypeWorkspace,
				},
			},
		},
		"only .bzl file": {
			files:   []string{"foo.bzl"},
			wantErr: true,
		},
		"all kinds": {
			files: []string{"WORKSPACE", "WORKSPACE.bzlmod", "foo.bzl", "bar.bzl", "unused.txt", "folder/baz.bzl"},
			wantFiles: []BazelFile{
				{
					RelPath: "WORKSPACE.bzlmod",
					AbsPath: "/WORKSPACE.bzlmod",
					Type:    BazelFileTypeWorkspace,
				},
				{
					RelPath: "foo.bzl",
					AbsPath: "/foo.bzl",
					Type:    BazelFileTypeBzl,
				},
				{
					RelPath: "bar.bzl",
					AbsPath: "/bar.bzl",
					Type:    BazelFileTypeBzl,
				},
				{
					RelPath: "folder/baz.bzl",
					AbsPath: "/folder/baz.bzl",
					Type:    BazelFileTypeBzl,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			fs := afero.NewMemMapFs()
			for _, file := range tc.files {
				_, err := fs.Create(file)
				assert.NoError(err)
			}

			helper := Helper{
				fs:            fs,
				workspaceRoot: "/",
			}
			gotFiles, err := helper.FindFiles()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantFiles, gotFiles)
		})
	}
}

func TestLoadFile(t *testing.T) {
	testCases := map[string]struct {
		file     BazelFile
		contents string
		wantErr  bool
	}{
		"file does not exist": {
			file: BazelFile{
				RelPath: "foo.bzl",
				AbsPath: "/foo.bzl",
				Type:    BazelFileTypeBzl,
			},
			wantErr: true,
		},
		"file has unknown type": {
			file: BazelFile{
				RelPath: "foo.txt",
				AbsPath: "/foo.txt",
				Type:    BazelFileType(999),
			},
			contents: "foo",
			wantErr:  true,
		},
		"file is a bzl file": {
			file: BazelFile{
				RelPath: "foo.bzl",
				AbsPath: "/foo.bzl",
				Type:    BazelFileTypeBzl,
			},
			contents: "load(\"bar.bzl\", \"bar\")",
		},
		"file is a workspace file": {
			file: BazelFile{
				RelPath: "WORKSPACE",
				AbsPath: "/WORKSPACE",
				Type:    BazelFileTypeWorkspace,
			},
			contents: "workspace(name = \"foo\")",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			fs := afero.NewMemMapFs()
			if tc.contents != "" {
				err := afero.WriteFile(fs, tc.file.RelPath, []byte(tc.contents), 0o644)
				require.NoError(err)
			}

			helper := Helper{
				fs:            fs,
				workspaceRoot: "/",
			}
			_, err := helper.LoadFile(tc.file)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestReadWriteFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "a.bzl", []byte("load(\"bar.bzl\", \"bar\")\n"), 0o644)
	require.NoError(err)
	helper := Helper{
		fs:            fs,
		workspaceRoot: "/",
	}
	bf, err := helper.LoadFile(BazelFile{
		RelPath: "a.bzl",
		AbsPath: "/a.bzl",
		Type:    BazelFileTypeBzl,
	})
	require.NoError(err)
	err = helper.WriteFile(BazelFile{
		RelPath: "b.bzl",
		AbsPath: "/b.bzl",
		Type:    BazelFileTypeBzl,
	}, bf)
	require.NoError(err)
	_, err = fs.Stat("b.bzl")
	assert.NoError(err)
	contents, err := afero.ReadFile(fs, "b.bzl")
	assert.NoError(err)
	assert.Equal("load(\"bar.bzl\", \"bar\")\n", string(contents))
}

func TestDiff(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "WORKSPACE.bzlmod", []byte(""), 0o644)
	require.NoError(err)
	helper := Helper{
		fs:            fs,
		workspaceRoot: "/",
	}
	fileRef := BazelFile{
		RelPath: "WORKSPACE.bzlmod",
		AbsPath: "/WORKSPACE.bzlmod",
		Type:    BazelFileTypeWorkspace,
	}
	bf, err := helper.LoadFile(fileRef)
	require.NoError(err)
	diff, err := helper.Diff(fileRef, bf)
	require.NoError(err)
	assert.Empty(diff)
	bf.Stmt = edit.InsertAtEnd(
		bf.Stmt,
		&build.CallExpr{
			X: &build.Ident{Name: "workspace"},
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{Value: "foo"},
				},
			},
		},
	)
	diff, err = helper.Diff(fileRef, bf)
	require.NoError(err)
	assert.Equal("--- a/WORKSPACE.bzlmod\n+++ b/WORKSPACE.bzlmod\n@@ -1 +1 @@\n+workspace(name = \"foo\")\n", diff)
	err = helper.WriteFile(fileRef, bf)
	require.NoError(err)
	contents, err := afero.ReadFile(fs, "WORKSPACE.bzlmod")
	assert.NoError(err)
	assert.Equal("workspace(name = \"foo\")\n", string(contents))
	diff, err = helper.Diff(fileRef, bf)
	require.NoError(err)
	assert.Empty(diff)
}
