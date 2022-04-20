package diskencryption

import (
	"errors"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenClose(t *testing.T) {
	testCases := map[string]struct {
		initByNameErr error
		operations    []string
		errExpected   bool
	}{
		"open and close work": {
			operations: []string{"open", "close"},
		},
		"opening twice fails": {
			operations:  []string{"open", "open"},
			errExpected: true,
		},
		"closing first fails": {
			operations:  []string{"close"},
			errExpected: true,
		},
		"initByName failure detected": {
			initByNameErr: errors.New("initByNameErr"),
			operations:    []string{"open"},
			errExpected:   true,
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
			if tc.errExpected {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestUUID(t *testing.T) {
	testCases := map[string]struct {
		open         bool
		expectedUUID string
		errExpected  bool
	}{
		"getting uuid works": {
			open:         true,
			expectedUUID: "uuid",
		},
		"getting uuid on closed device fails": {
			errExpected: true,
		},
		"empty uuid is detected": {
			open:        true,
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			crypt := Cryptsetup{
				fs: afero.NewMemMapFs(),
				initByName: func(name string) (cryptdevice, error) {
					return &stubCryptdevice{uuid: tc.expectedUUID}, nil
				},
			}

			if tc.open {
				require.NoError(crypt.Open())
				defer crypt.Close()
			}
			uuid, err := crypt.UUID()
			if tc.errExpected {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedUUID, uuid)
		})
	}
}

func TestUpdatePassphrase(t *testing.T) {
	testCases := map[string]struct {
		writePassphrase              bool
		open                         bool
		keyslotChangeByPassphraseErr error
		errExpected                  bool
	}{
		"updating passphrase works": {
			writePassphrase: true,
			open:            true,
		},
		"updating passphrase on closed device fails": {
			errExpected: true,
		},
		"reading initial passphrase can fail": {
			open:        true,
			errExpected: true,
		},
		"changing keyslot passphrase can fail": {
			open:                         true,
			writePassphrase:              true,
			keyslotChangeByPassphraseErr: errors.New("keyslotChangeByPassphraseErr"),
			errExpected:                  true,
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
			if tc.errExpected {
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

func (s *stubCryptdevice) KeyslotChangeByPassphrase(currentKeyslot int, newKeyslot int, currentPassphrase string, newPassphrase string) error {
	return s.keyslotChangeErr
}

func (s *stubCryptdevice) Free() bool {
	return false
}
