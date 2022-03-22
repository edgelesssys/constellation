package deploy

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/debugd/debugd/deploy/passwd"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLinuxUser(t *testing.T) {
	username := "user"

	testCases := map[string]struct {
		userCreator    *stubUserCreator
		passwdContents string
		expectErr      bool
		expectedUser   LinuxUser
	}{
		"get works": {
			userCreator:    &stubUserCreator{},
			passwdContents: "user:x:1000:1000:user:/home/user:/bin/bash\n",
			expectErr:      false,
			expectedUser: LinuxUser{
				Username: "user",
				Home:     "/home/user",
				Uid:      1000,
				Gid:      1000,
			},
		},
		"user does not exist": {
			userCreator:    &stubUserCreator{},
			passwdContents: "",
			expectErr:      true,
		},
		"parse fails": {
			userCreator:    &stubUserCreator{},
			passwdContents: "invalid contents\n",
			expectErr:      true,
		},
		"invalid uid": {
			userCreator:    &stubUserCreator{},
			passwdContents: "user:x:invalid:1000:user:/home/user:/bin/bash\n",
			expectErr:      true,
		},
		"invalid gid": {
			userCreator:    &stubUserCreator{},
			passwdContents: "user:x:1000:invalid:user:/home/user:/bin/bash\n",
			expectErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			assert.NoError(afero.WriteFile(fs, "/etc/passwd", []byte(tc.passwdContents), 0o755))
			manager := LinuxUserManager{
				fs:      fs,
				passwd:  passwd.Passwd{},
				creator: tc.userCreator,
			}
			user, err := manager.getLinuxUser(username)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedUser, user)
		})
	}
}

func TestEnsureLinuxUserExists(t *testing.T) {
	username := "user"

	testCases := map[string]struct {
		userCreator    *stubUserCreator
		passwdContents string
		expectErr      bool
		expectedUser   LinuxUser
	}{
		"create works": {
			userCreator:    &stubUserCreator{},
			passwdContents: "user:x:1000:1000:user:/home/user:/bin/bash\n",
			expectErr:      false,
			expectedUser: LinuxUser{
				Username: "user",
				Home:     "/home/user",
				Uid:      1000,
				Gid:      1000,
			},
		},
		"create fails": {
			userCreator: &stubUserCreator{
				createUserErr: errors.New("create fails"),
			},
			passwdContents: "user:x:1000:1000:user:/home/user:/bin/bash\n",
			expectErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			assert.NoError(afero.WriteFile(fs, "/etc/passwd", []byte(tc.passwdContents), 0o755))
			manager := LinuxUserManager{
				fs:      fs,
				passwd:  passwd.Passwd{},
				creator: tc.userCreator,
			}
			user, err := manager.EnsureLinuxUserExists(context.Background(), username)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedUser, user)
			assert.ElementsMatch([]string{username}, tc.userCreator.usernames)
		})
	}
}

type stubUserCreator struct {
	usernames     []string
	createUserErr error
}

func (s *stubUserCreator) CreateUser(ctx context.Context, username string) error {
	s.usernames = append(s.usernames, username)
	return s.createUserErr
}
