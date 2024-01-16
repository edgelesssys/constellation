/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package secureboot

import (
	"encoding/base64"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestDatabaseFromFile(t *testing.T) {
	testCases := map[string]struct {
		pk      string
		keks    []string
		dbs     []string
		wantErr bool
	}{
		"found": {
			pk:   "pki/pk.cer",
			keks: []string{"pki/kek.cer"},
			dbs:  []string{"pki/db.cer"},
		},
		"pk not found": {
			pk:      "pki/missing",
			keks:    []string{"pki/kek.cer"},
			dbs:     []string{"pki/db.cer"},
			wantErr: true,
		},
		"kek not found": {
			pk:      "pki/pk.cer",
			keks:    []string{"pki/missing"},
			dbs:     []string{"pki/db.cer"},
			wantErr: true,
		},
		"db not found": {
			pk:      "pki/pk.cer",
			keks:    []string{"pki/kek.cer"},
			dbs:     []string{"pki/missing"},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			fs := afero.NewMemMapFs()
			require.NoError(afero.WriteFile(fs, "pki/pk.cer", []byte("pk"), 0o644))
			require.NoError(afero.WriteFile(fs, "pki/kek.cer", []byte("kek"), 0o644))
			require.NoError(afero.WriteFile(fs, "pki/db.cer", []byte("db"), 0o644))
			db, err := DatabaseFromFiles(fs, tc.pk, tc.keks, tc.dbs)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(db.Keks, 1)
			assert.Len(db.DBs, 1)
			assert.Equal("pk", string(db.PK))
			assert.Equal("kek", string(db.Keks[0]))
			assert.Equal("db", string(db.DBs[0]))
		})
	}
}

func TestVarsStoreFromFiles(t *testing.T) {
	testCases := map[string]struct {
		pk      string
		kek     string
		db      string
		dbx     string
		wantErr bool
	}{
		"found": {
			pk:  "pki/pk.esl",
			kek: "pki/kek.esl",
			db:  "pki/db.esl",
			dbx: "pki/dbx.esl",
		},
		"no dbx": {
			pk:  "pki/pk.esl",
			kek: "pki/kek.esl",
			db:  "pki/db.esl",
		},
		"pk not found": {
			pk:      "pki/missing",
			kek:     "pki/kek.esl",
			db:      "pki/db.esl",
			wantErr: true,
		},
		"kek not found": {
			pk:      "pki/pk.esl",
			kek:     "pki/missing",
			db:      "pki/db.esl",
			wantErr: true,
		},
		"db not found": {
			pk:      "pki/pk.esl",
			kek:     "pki/kek.esl",
			db:      "pki/missing",
			wantErr: true,
		},
		"dbx not found": {
			pk:      "pki/pk.esl",
			kek:     "pki/kek.esl",
			db:      "pki/db.esl",
			dbx:     "pki/missing",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			fs := afero.NewMemMapFs()
			require.NoError(afero.WriteFile(fs, "pki/pk.esl", []byte("pk"), 0o644))
			require.NoError(afero.WriteFile(fs, "pki/kek.esl", []byte("kek"), 0o644))
			require.NoError(afero.WriteFile(fs, "pki/db.esl", []byte("db"), 0o644))
			require.NoError(afero.WriteFile(fs, "pki/dbx.esl", []byte("dbx"), 0o644))
			store, err := VarStoreFromFiles(fs, tc.pk, tc.kek, tc.db, tc.dbx)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			if tc.dbx == "" {
				assert.Len(store, 3)
			} else {
				assert.Len(store, 4)
				assert.Equal("dbx", store[3].Name)
				assert.Equal("dbx", string(store[3].Data))
				assert.Equal(secureDatabaseGUID, store[3].GUID)
			}
			assert.Equal("PK", store[0].Name)
			assert.Equal("pk", string(store[0].Data))
			assert.Equal(globalEFIGUID, store[0].GUID)
			assert.Equal("KEK", store[1].Name)
			assert.Equal("kek", string(store[1].Data))
			assert.Equal(globalEFIGUID, store[1].GUID)
			assert.Equal("db", store[2].Name)
			assert.Equal("db", string(store[2].Data))
			assert.Equal(secureDatabaseGUID, store[2].GUID)
		})
	}
}

func TestToAWS(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	fs := afero.NewMemMapFs()
	require.NoError(afero.WriteFile(fs, "pki/pk.esl", []byte("pk"), 0o644))
	require.NoError(afero.WriteFile(fs, "pki/kek.esl", []byte("kek"), 0o644))
	require.NoError(afero.WriteFile(fs, "pki/db.esl", []byte("db"), 0o644))
	require.NoError(afero.WriteFile(fs, "pki/dbx.esl", []byte("dbx"), 0o644))
	store, err := VarStoreFromFiles(fs, "pki/pk.esl", "pki/kek.esl", "pki/db.esl", "pki/dbx.esl")
	require.NoError(err)
	awsData, err := store.ToAWS()
	require.NoError(err)
	out, err := base64.StdEncoding.DecodeString(awsData)
	assert.NoError(err)
	assert.Equal([]byte{
		0x41, 0x4d, 0x5a, 0x4e, 0x55, 0x45, 0x46, 0x49, 0x5d, 0x52, 0x8c, 0xf0, 0x00, 0x00, 0x00, 0x00,
		0x78, 0xf9, 0x6b, 0xb7, 0xd9, 0xf7, 0x62, 0x81, 0xda, 0xc4, 0x04, 0xa5, 0x03, 0xbc, 0x61, 0xac,
		0x82, 0x6c, 0x74, 0x5b, 0xd5, 0xb1, 0xb8, 0x50, 0x81, 0xc8, 0x25, 0xbe, 0xb0, 0x69, 0x3d, 0x6f,
		0x57, 0x6f, 0x18, 0x33, 0x3b, 0x95, 0xaa, 0x36, 0xc0, 0xdc, 0x9d, 0x92, 0x84, 0x60, 0xa1, 0xb7,
		0xcc, 0xa9, 0xe1, 0x83, 0x94, 0xa4, 0x0a, 0x24, 0x26, 0x35, 0x6d, 0x00, 0x04, 0x00, 0x00, 0xff,
		0xff, 0xd6, 0x50, 0x28, 0x9b,
	}, out)
}
