/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package ssh

import (
	"context"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/deploy/user"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGetAuthorizedKeys(t *testing.T) {
	testCases := map[string]struct {
		authorized map[UserKey]bool
		want       []UserKey
	}{
		"success": {
			authorized: map[UserKey]bool{
				{Username: "user1", PublicKey: "ssh-rsa test1=="}: true,
				{Username: "user2", PublicKey: "ssh-rsa test2=="}: true,
			},
			want: []UserKey{
				{Username: "user1", PublicKey: "ssh-rsa test1=="},
				{Username: "user2", PublicKey: "ssh-rsa test2=="},
			},
		},
		"empty": {
			authorized: map[UserKey]bool{},
			want:       []UserKey(nil),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			sshAccess := Access{authorized: tc.authorized}

			keys := sshAccess.GetAuthorizedKeys()
			assert.ElementsMatch(tc.want, keys)
		})
	}
}

func TestDeployAuthorizedKey(t *testing.T) {
	authorizedKey := UserKey{
		Username:  "user",
		PublicKey: "ssh-rsa testkey",
	}

	testCases := map[string]struct {
		fs               afero.Fs
		passwdContents   string
		alreadyDeployed  bool
		readonly         bool
		wantErr          bool
		wantFile         bool
		wantFileContents string
	}{
		"deploy works": {
			fs:               afero.NewMemMapFs(),
			wantErr:          false,
			wantFile:         true,
			wantFileContents: "ssh-rsa testkey\n",
		},
		"appending ssh key works": {
			fs:               memMapFsWithFile("/var/home/user/.ssh/authorized_keys.d/constellation-ssh-keys", "ssh-rsa preexistingkey\n"),
			wantErr:          false,
			wantFile:         true,
			wantFileContents: "ssh-rsa preexistingkey\nssh-rsa testkey\n",
		},
		"redeployment avoided": {
			fs:              afero.NewMemMapFs(),
			wantErr:         false,
			alreadyDeployed: true,
			wantFile:        false,
		},
		"readonly fs": {
			fs:       afero.NewMemMapFs(),
			readonly: true,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			userManager := user.NewLinuxUserManagerFake(tc.fs)

			assert.NoError(afero.WriteFile(userManager.Fs, "/etc/passwd", []byte(tc.passwdContents), 0o755))
			if tc.readonly {
				userManager.Fs = afero.NewReadOnlyFs(userManager.Fs)
			}
			authorized := map[UserKey]bool{}
			if tc.alreadyDeployed {
				authorized[UserKey{
					Username:  "user",
					PublicKey: "ssh-rsa testkey",
				}] = true
			}
			sshAccess := Access{
				log:         logger.NewTest(t),
				userManager: userManager,
				mux:         sync.Mutex{},
				authorized:  authorized,
			}
			err := sshAccess.DeployAuthorizedKey(context.Background(), authorizedKey)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			if tc.wantFile {
				fileContents, err := afero.ReadFile(userManager.Fs, "/var/home/user/.ssh/authorized_keys.d/constellation-ssh-keys")
				assert.NoError(err)
				assert.Equal(tc.wantFileContents, string(fileContents))
			} else {
				exists, err := afero.Exists(userManager.Fs, "/var/home/user/.ssh/authorized_keys.d/constellation-ssh-keys")
				assert.NoError(err)
				assert.False(exists)
			}
		})
	}
}

func memMapFsWithFile(path string, contents string) afero.Fs {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, path, []byte(contents), 0o755)
	if err != nil {
		panic(err)
	}
	return fs
}
