/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package diskencryption

import (
	"errors"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestOpenClose(t *testing.T) {
	testCases := map[string]struct {
		initByNameErr error
		operations    []string
		wantErr       bool
	}{
		"open and close work": {
			operations: []string{"open", "close"},
		},
		"opening twice fails": {
			operations: []string{"open", "open"},
			wantErr:    true,
		},
		"closing first fails": {
			operations: []string{"close"},
			wantErr:    true,
		},
		"initByName failure detected": {
			initByNameErr: errors.New("initByNameErr"),
			operations:    []string{"open"},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			crypt := Cryptsetup{
				fs: afero.NewMemMapFs(),
				initByName: func(name string) (cryptdevice, error) {
					return &stubCryptdevice{}, tc.initByNameErr
				},
			}

			err := executeOperations(&crypt, tc.operations)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestUUID(t *testing.T) {
	testCases := map[string]struct {
		open     bool
		wantUUID string
		wantErr  bool
	}{
		"getting uuid works": {
			open:     true,
			wantUUID: "uuid",
		},
		"getting uuid on closed device fails": {
			wantErr: true,
		},
		"empty uuid is detected": {
			open:    true,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			crypt := Cryptsetup{
				fs: afero.NewMemMapFs(),
				initByName: func(name string) (cryptdevice, error) {
					return &stubCryptdevice{uuid: tc.wantUUID}, nil
				},
			}

			if tc.open {
				require.NoError(crypt.Open())
				defer crypt.Close()
			}
			uuid, err := crypt.UUID()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantUUID, uuid)
		})
	}
}

func TestUpdatePassphrase(t *testing.T) {
	testCases := map[string]struct {
		writePassphrase              bool
		open                         bool
		keyslotChangeByPassphraseErr error
		wantErr                      bool
	}{
		"updating passphrase works": {
			writePassphrase: true,
			open:            true,
		},
		"updating passphrase on closed device fails": {
			wantErr: true,
		},
		"reading initial passphrase can fail": {
			open:    true,
			wantErr: true,
		},
		"changing keyslot passphrase can fail": {
			open:                         true,
			writePassphrase:              true,
			keyslotChangeByPassphraseErr: errors.New("keyslotChangeByPassphraseErr"),
			wantErr:                      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			if tc.writePassphrase {
				require.NoError(fs.MkdirAll(path.Base(initialKeyPath), 0o777))
				require.NoError(afero.WriteFile(fs, initialKeyPath, []byte("key"), 0o777))
			}

			crypt := Cryptsetup{
				fs: fs,
				initByName: func(name string) (cryptdevice, error) {
					return &stubCryptdevice{keyslotChangeErr: tc.keyslotChangeByPassphraseErr}, nil
				},
			}

			if tc.open {
				require.NoError(crypt.Open())
				defer crypt.Close()
			}
			err := crypt.UpdatePassphrase("new-key")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func executeOperations(crypt *Cryptsetup, operations []string) error {
	for _, operation := range operations {
		var err error
		switch operation {
		case "open":
			err = crypt.Open()
		case "close":
			err = crypt.Close()
		default:
			panic("unknown operation")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type stubCryptdevice struct {
	uuid             string
	keyslotChangeErr error
}

func (s *stubCryptdevice) GetUUID() string {
	return s.uuid
}

func (s *stubCryptdevice) KeyslotChangeByPassphrase(_, _ int, _, _ string) error {
	return s.keyslotChangeErr
}

func (s *stubCryptdevice) Free() bool {
	return false
}
