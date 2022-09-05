/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package user

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	filename := "/etc/passwd"

	testCases := map[string]struct {
		passwdContents string
		createFile     bool
		wantEntries    Entries
		wantErr        bool
	}{
		"parse works": {
			passwdContents: "root:x:0:0:root:/root:/bin/bash\n",
			createFile:     true,
			wantEntries: Entries{
				"root": {
					Password:  "x",
					UID:       "0",
					GID:       "0",
					GECOS:     "root",
					Directory: "/root",
					Shell:     "/bin/bash",
				},
			},
			wantErr: false,
		},
		"multiple lines": {
			passwdContents: "root:x:0:0:root:/root:/bin/bash\nfoo:y:1:2:bar:baz:sh",
			createFile:     true,
			wantEntries: Entries{
				"root": {
					Password:  "x",
					UID:       "0",
					GID:       "0",
					GECOS:     "root",
					Directory: "/root",
					Shell:     "/bin/bash",
				},
				"foo": {
					Password:  "y",
					UID:       "1",
					GID:       "2",
					GECOS:     "bar",
					Directory: "baz",
					Shell:     "sh",
				},
			},
			wantErr: false,
		},
		"passwd is corrupt": {
			passwdContents: "too:few:fields\n",
			createFile:     true,
			wantErr:        true,
		},
		"file does not exist": {
			createFile: false,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			if tc.createFile {
				assert.NoError(afero.WriteFile(fs, filename, []byte(tc.passwdContents), 0o644))
			}
			passwd := Passwd{}
			entries, err := passwd.Parse(fs)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantEntries, entries)
		})
	}
}
