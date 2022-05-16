package user

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLinuxUser(t *testing.T) {
	username := "user"

	testCases := map[string]struct {
		passwdContents string
		wantErr        bool
		wantUser       LinuxUser
	}{
		"get works": {
			passwdContents: "user:x:1000:1000:user:/home/user:/bin/bash\n",
			wantErr:        false,
			wantUser: LinuxUser{
				Username: "user",
				Home:     "/home/user",
				Uid:      1000,
				Gid:      1000,
			},
		},
		"user does not exist": {
			passwdContents: "",
			wantErr:        true,
		},
		"parse fails": {
			passwdContents: "invalid contents\n",
			wantErr:        true,
		},
		"invalid uid": {
			passwdContents: "user:x:invalid:1000:user:/home/user:/bin/bash\n",
			wantErr:        true,
		},
		"invalid gid": {
			passwdContents: "user:x:1000:invalid:user:/home/user:/bin/bash\n",
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			assert.NoError(afero.WriteFile(fs, "/etc/passwd", []byte(tc.passwdContents), 0o755))
			manager := NewLinuxUserManagerFake(fs)
			user, err := manager.getLinuxUser(username)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantUser, user)
		})
	}
}

func TestEnsureLinuxUserExists(t *testing.T) {
	username := "user"

	testCases := map[string]struct {
		userCreator *StubUserCreator
		wantErr     bool
		wantUser    LinuxUser
	}{
		"create works": {
			userCreator: &StubUserCreator{},
			wantErr:     false,
			wantUser: LinuxUser{
				Username: "user",
				Home:     "/home/user",
				Uid:      1000,
				Gid:      1000,
			},
		},
		"create fails": {
			userCreator: &StubUserCreator{
				createUserErr: errors.New("create fails"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			manager := NewLinuxUserManagerFake(fs)
			tc.userCreator.fs = fs
			manager.Creator = tc.userCreator
			user, err := manager.EnsureLinuxUserExists(context.Background(), username)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantUser, user)
			assert.ElementsMatch([]string{username}, tc.userCreator.usernames)
		})
	}
}

func TestStringInSlice(t *testing.T) {
	assert := assert.New(t)
	testSlice := []string{"abc", "efg", "xyz"}

	assert.True(stringInSlice("efg", testSlice))
	assert.False(stringInSlice("hij", testSlice))
}
