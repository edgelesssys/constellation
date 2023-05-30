/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package file

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"gopkg.in/yaml.v3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestWrite(t *testing.T) {
	testCases := map[string]struct {
		fs              afero.Fs
		setupFs         func(af afero.Afero) error
		name            string
		content         string
		expectedContent string
		options         Option
		wantErr         bool
		wantAppend      bool
	}{
		"successful write": {
			fs:              afero.NewMemMapFs(),
			content:         "asdf",
			expectedContent: "asdf",
			name:            "somedir/somefile",
		},
		"successful overwrite": {
			fs:              afero.NewMemMapFs(),
			setupFs:         func(af afero.Afero) error { return af.WriteFile("somedir/somefile", []byte{}, 0o644) },
			content:         "asdf",
			expectedContent: "asdf",
			name:            "somedir/somefile",
			options:         OptOverwrite,
		},
		"successful append": {
			fs:              afero.NewMemMapFs(),
			setupFs:         func(af afero.Afero) error { return af.WriteFile("somedir/somefile", []byte("fdsa"), 0o644) },
			content:         "asdf",
			expectedContent: "fdsaasdf",
			name:            "somedir/somefile",
			options:         OptAppend,
		},
		"read only fs": {
			fs:      afero.NewReadOnlyFs(afero.NewMemMapFs()),
			name:    "somedir/somefile",
			wantErr: true,
		},
		"file already exists": {
			fs:      afero.NewMemMapFs(),
			setupFs: func(af afero.Afero) error { return af.WriteFile("somedir/somefile", []byte{}, 0o644) },
			name:    "somedir/somefile",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := NewHandler(tc.fs)
			if tc.setupFs != nil {
				require.NoError(tc.setupFs(afero.Afero{Fs: tc.fs}))
			}

			if tc.wantErr {
				assert.Error(handler.Write(tc.name, []byte(tc.content), tc.options))
			} else {
				assert.NoError(handler.Write(tc.name, []byte(tc.content), tc.options))
				content, err := handler.Read(tc.name)
				require.NoError(err)
				assert.Equal(tc.expectedContent, string(content))
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	type testContent struct {
		First  string
		Second int
	}
	someContent := testContent{
		First:  "first",
		Second: 2,
	}
	jsonContent, err := json.MarshalIndent(someContent, "", "\t")
	require.NoError(t, err)

	testCases := map[string]struct {
		fs          afero.Fs
		setupFs     func(fs *afero.Afero) error
		name        string
		wantContent any
		wantErr     bool
	}{
		"successful read": {
			fs:          afero.NewMemMapFs(),
			name:        "test/statefile",
			setupFs:     func(fs *afero.Afero) error { return fs.WriteFile("test/statefile", jsonContent, 0o755) },
			wantContent: someContent,
		},
		"file not existent": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			wantErr: true,
		},
		"file not json": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			setupFs: func(fs *afero.Afero) error { return fs.WriteFile("test/statefile", []byte{0x1}, 0o755) },
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := NewHandler(tc.fs)
			if tc.setupFs != nil {
				require.NoError(tc.setupFs(handler.fs))
			}

			resultContent := &testContent{}
			if tc.wantErr {
				assert.Error(handler.ReadJSON(tc.name, resultContent))
			} else {
				assert.NoError(handler.ReadJSON(tc.name, resultContent))
				assert.Equal(tc.wantContent, *resultContent)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	type testContent struct {
		First  string
		Second int
	}
	someContent := testContent{
		First:  "first",
		Second: 2,
	}
	notMarshalableContent := struct{ Foo chan int }{Foo: make(chan int)}

	testCases := map[string]struct {
		fs      afero.Fs
		setupFs func(af afero.Afero) error
		name    string
		content any
		options Option
		wantErr bool
	}{
		"successful write": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			content: someContent,
		},
		"successful overwrite": {
			fs:      afero.NewMemMapFs(),
			setupFs: func(af afero.Afero) error { return af.WriteFile("test/statefile", []byte{}, 0o644) },
			name:    "test/statefile",
			content: someContent,
			options: OptOverwrite,
		},
		"read only fs": {
			fs:      afero.NewReadOnlyFs(afero.NewMemMapFs()),
			name:    "test/statefile",
			content: someContent,
			wantErr: true,
		},
		"file already exists": {
			fs:      afero.NewMemMapFs(),
			setupFs: func(af afero.Afero) error { return af.WriteFile("test/statefile", []byte{}, 0o644) },
			name:    "test/statefile",
			content: someContent,
			wantErr: true,
		},
		"marshal error": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			content: notMarshalableContent,
			wantErr: true,
		},
		"mkdirAll works": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			content: someContent,
			options: OptMkdirAll,
		},
		// TODO: add tests for mkdirAll actually creating the necessary folders when https://github.com/spf13/afero/issues/270 is fixed.
		// Currently, MemMapFs will create files in nonexistent directories due to a bug in afero,
		// making it impossible to test the actual behavior of the mkdirAll parameter.
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := NewHandler(tc.fs)
			if tc.setupFs != nil {
				require.NoError(tc.setupFs(afero.Afero{Fs: tc.fs}))
			}

			if tc.wantErr {
				assert.Error(handler.WriteJSON(tc.name, tc.content, tc.options))
			} else {
				assert.NoError(handler.WriteJSON(tc.name, tc.content, tc.options))
				resultContent := &testContent{}
				assert.NoError(handler.ReadJSON(tc.name, resultContent))
				assert.Equal(tc.content, *resultContent)
			}
		})
	}
}

func TestReadYAML(t *testing.T) {
	type testContent struct {
		First  string
		Second int
	}
	someContent := testContent{
		First:  "first",
		Second: 2,
	}
	yamlContent, err := yaml.Marshal(someContent)
	require.NoError(t, err)

	testCases := map[string]struct {
		fs          afero.Fs
		setupFs     func(fs *afero.Afero) error
		name        string
		wantContent any
		wantErr     bool
	}{
		"successful read": {
			fs:          afero.NewMemMapFs(),
			name:        "test/config.yaml",
			setupFs:     func(fs *afero.Afero) error { return fs.WriteFile("test/config.yaml", yamlContent, 0o755) },
			wantContent: someContent,
		},
		"file not existent": {
			fs:      afero.NewMemMapFs(),
			name:    "test/config.yaml",
			wantErr: true,
		},
		"file not yaml": {
			fs:      afero.NewMemMapFs(),
			name:    "test/config.yaml",
			setupFs: func(fs *afero.Afero) error { return fs.WriteFile("test/config.yaml", []byte{0x1}, 0o755) },
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := NewHandler(tc.fs)
			if tc.setupFs != nil {
				require.NoError(tc.setupFs(handler.fs))
			}

			resultContent := &testContent{}
			if tc.wantErr {
				assert.Error(handler.ReadYAML(tc.name, resultContent))
			} else {
				assert.NoError(handler.ReadYAML(tc.name, resultContent))
				assert.Equal(tc.wantContent, *resultContent)
			}
		})
	}
}

func TestReadYAMLStrictUnknownFieldFails(t *testing.T) {
	assert := assert.New(t)

	type SampleConfig struct {
		Version string `yaml:"version"`
		Value   string `yaml:"value"`
	}
	yamlConfig := `
	version: "1.0.0"
	value: "foobar"
	sneakyValue: "superSecret"
	`

	handler := NewHandler(afero.NewMemMapFs())
	assert.NoError(handler.Write(constants.ConfigFilename, []byte(yamlConfig), OptNone))

	var readInConfig SampleConfig
	assert.Error(handler.ReadYAMLStrict(constants.ConfigFilename, &readInConfig))
}

func TestWriteYAML(t *testing.T) {
	type testContent struct {
		First  string
		Second int
	}
	someContent := testContent{
		First:  "first",
		Second: 2,
	}
	notMarshalableContent := struct{ Foo chan int }{Foo: make(chan int)}

	testCases := map[string]struct {
		fs      afero.Fs
		setupFs func(af afero.Afero) error
		name    string
		content any
		options Option
		wantErr bool
	}{
		"successful write": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			content: someContent,
		},
		"successful overwrite": {
			fs:      afero.NewMemMapFs(),
			setupFs: func(af afero.Afero) error { return af.WriteFile("test/statefile", []byte{}, 0o644) },
			name:    "test/statefile",
			content: someContent,
			options: OptOverwrite,
		},
		"read only fs": {
			fs:      afero.NewReadOnlyFs(afero.NewMemMapFs()),
			name:    "test/statefile",
			content: someContent,
			wantErr: true,
		},
		"file already exists": {
			fs:      afero.NewMemMapFs(),
			setupFs: func(af afero.Afero) error { return af.WriteFile("test/statefile", []byte{}, 0o644) },
			name:    "test/statefile",
			content: someContent,
			wantErr: true,
		},
		"marshal error": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			content: notMarshalableContent,
			wantErr: true,
		},
		"mkdirAll works": {
			fs:      afero.NewMemMapFs(),
			name:    "test/statefile",
			content: someContent,
			options: OptMkdirAll,
		},
		// TODO: add tests for mkdirAll actually creating the necessary folders when https://github.com/spf13/afero/issues/270 is fixed.
		// Currently, MemMapFs will create files in nonexistent directories due to a bug in afero,
		// making it impossible to test the actual behavior of the mkdirAll parameter.
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := NewHandler(tc.fs)
			if tc.setupFs != nil {
				require.NoError(tc.setupFs(afero.Afero{Fs: tc.fs}))
			}

			if tc.wantErr {
				assert.Error(handler.WriteYAML(tc.name, tc.content, tc.options))
			} else {
				assert.NoError(handler.WriteYAML(tc.name, tc.content, tc.options))
				resultContent := &testContent{}
				assert.NoError(handler.ReadYAML(tc.name, resultContent))
				assert.Equal(tc.content, *resultContent)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.NewMemMapFs()
	handler := NewHandler(fs)
	aferoHelper := afero.Afero{Fs: fs}
	require.NoError(aferoHelper.WriteFile("a", []byte{0xa}, 0o644))
	require.NoError(aferoHelper.WriteFile("b", []byte{0xb}, 0o644))
	require.NoError(aferoHelper.WriteFile("c", []byte{0xc}, 0o644))

	assert.NoError(handler.Remove("a"))
	assert.NoError(handler.Remove("b"))
	assert.NoError(handler.Remove("c"))

	_, err := handler.fs.Stat("a")
	assert.ErrorIs(err, afero.ErrFileNotFound)
	_, err = handler.fs.Stat("b")
	assert.ErrorIs(err, afero.ErrFileNotFound)
	_, err = handler.fs.Stat("c")
	assert.ErrorIs(err, afero.ErrFileNotFound)

	assert.Error(handler.Remove("d"))
}

func TestCopyFile(t *testing.T) {
	perms := fs.FileMode(0o644)

	setupFs := func(existingFiles ...string) afero.Fs {
		fs := afero.NewMemMapFs()
		aferoHelper := afero.Afero{Fs: fs}
		for _, file := range existingFiles {
			require.NoError(t, aferoHelper.WriteFile(file, []byte{}, perms))
		}
		return fs
	}

	testCases := map[string]struct {
		fs         afero.Fs
		copyFiles  [][]string
		checkFiles []string
		opts       []Option
		wantErr    bool
	}{
		"successful copy": {
			fs:         setupFs("a"),
			copyFiles:  [][]string{{"a", "b"}},
			checkFiles: []string{"b"},
		},
		"copy to existing file overwrite": {
			fs:         setupFs("a", "b"),
			copyFiles:  [][]string{{"a", "b"}},
			checkFiles: []string{"b"},
			opts:       []Option{OptOverwrite},
		},
		"copy to existing file no overwrite": {
			fs:         setupFs("a", "b"),
			copyFiles:  [][]string{{"a", "b"}},
			checkFiles: []string{"b"},
			wantErr:    true,
		},
		"file doesn't exist": {
			fs:         setupFs("a"),
			copyFiles:  [][]string{{"b", "c"}},
			checkFiles: []string{"a"},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := NewHandler(tc.fs)
			for _, files := range tc.copyFiles {
				err := handler.CopyFile(files[0], files[1], tc.opts...)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
				}
			}

			for _, file := range tc.checkFiles {
				info, err := handler.fs.Stat(file)
				assert.Equal(perms, info.Mode())
				require.NoError(err)
			}
		})
	}
}

func TestCopyDir(t *testing.T) {
	setupHandler := func(existingFiles ...string) Handler {
		fs := afero.NewMemMapFs()
		handler := NewHandler(fs)
		for _, file := range existingFiles {
			err := handler.Write(file, []byte("some content"), OptMkdirAll)
			require.NoError(t, err)
		}
		return handler
	}

	testCases := map[string]struct {
		handler    Handler
		copyFiles  [][]string
		checkFiles []string
		opts       []Option
	}{
		"successful copy": {
			handler:    setupHandler(filepath.Join("someDir", "someFile"), filepath.Join("someDir", "someOtherDir", "someOtherFile")),
			copyFiles:  [][]string{{"someDir", "copiedDir"}},
			checkFiles: []string{filepath.Join("copiedDir", "someFile"), filepath.Join("copiedDir", "someOtherDir", "someOtherFile")},
		},
		"copy file": {
			handler:    setupHandler("someFile"),
			copyFiles:  [][]string{{"someFile", "copiedFile"}},
			checkFiles: []string{"copiedFile"},
		},
		"copy to existing dir overwrite": {
			handler:    setupHandler(filepath.Join("someDir", "someFile"), filepath.Join("someDir", "someOtherDir", "someOtherFile"), filepath.Join("copiedDir", "someExistingFile")),
			copyFiles:  [][]string{{"someDir", "copiedDir"}},
			checkFiles: []string{filepath.Join("copiedDir", "someFile"), filepath.Join("copiedDir", "someOtherDir", "someOtherFile"), filepath.Join("copiedDir", "someExistingFile")},
			opts:       []Option{OptOverwrite},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			for _, files := range tc.copyFiles {
				err := tc.handler.CopyDir(files[0], files[1], tc.opts...)
				require.NoError(err)
			}

			for _, file := range tc.checkFiles {
				_, err := tc.handler.fs.Stat(file)
				require.NoError(err)
			}
		})
	}
}
