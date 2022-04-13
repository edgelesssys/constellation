package file

import (
	"encoding/json"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		wantContent interface{}
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
		content interface{}
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
