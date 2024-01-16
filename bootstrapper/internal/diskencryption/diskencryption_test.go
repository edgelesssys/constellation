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
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
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

			crypt := DiskEncryption{
				fs:     fs,
				device: &stubCryptdevice{keyslotChangeErr: tc.keyslotChangeByPassphraseErr},
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

type stubCryptdevice struct {
	uuid             string
	uuidErr          error
	keyslotChangeErr error
}

func (s *stubCryptdevice) InitByName(_ string) (func(), error) {
	return func() {}, nil
}

func (s *stubCryptdevice) GetUUID() (string, error) {
	return s.uuid, s.uuidErr
}

func (s *stubCryptdevice) KeyslotChangeByPassphrase(_, _ int, _, _ string) error {
	return s.keyslotChangeErr
}

func (s *stubCryptdevice) SetConstellationStateDiskToken(bool) error {
	return nil
}
