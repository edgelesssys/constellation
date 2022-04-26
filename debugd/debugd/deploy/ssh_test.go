package deploy

import (
	"context"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/debugd/debugd/deploy/passwd"
	"github.com/edgelesssys/constellation/debugd/ssh"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploySSHAuthorizedKey(t *testing.T) {
	authorizedKey := ssh.SSHKey{
		Username: "user",
		KeyValue: "ssh-rsa testkey",
	}

	testCases := map[string]struct {
		fs               afero.Fs
		userCreator      *stubUserCreator
		passwdContents   string
		alreadyDeployed  bool
		readonly         bool
		wantErr          bool
		wantFile         bool
		wantFileContents string
	}{
		"deploy works": {
			fs:               afero.NewMemMapFs(),
			userCreator:      &stubUserCreator{},
			passwdContents:   "user:x:1000:1000:user:/home/user:/bin/bash\n",
			wantErr:          false,
			wantFile:         true,
			wantFileContents: "ssh-rsa testkey user\n",
		},
		"appending ssh key works": {
			fs:               memMapFsWithFile("/home/user/.ssh/authorized_keys.d/debugd", "ssh-rsa preexistingkey user\n"),
			userCreator:      &stubUserCreator{},
			passwdContents:   "user:x:1000:1000:user:/home/user:/bin/bash\n",
			wantErr:          false,
			wantFile:         true,
			wantFileContents: "ssh-rsa preexistingkey user\nssh-rsa testkey user\n",
		},
		"redeployment avoided": {
			fs:              afero.NewMemMapFs(),
			userCreator:     &stubUserCreator{},
			passwdContents:  "user:x:1000:1000:user:/home/user:/bin/bash\n",
			wantErr:         false,
			alreadyDeployed: true,
			wantFile:        false,
		},
		"user does not exist": {
			fs:             afero.NewMemMapFs(),
			userCreator:    &stubUserCreator{},
			passwdContents: "",
			wantErr:        true,
		},
		"readonly fs": {
			fs:             afero.NewMemMapFs(),
			userCreator:    &stubUserCreator{},
			passwdContents: "user:x:1000:1000:user:/home/user:/bin/bash\n",
			readonly:       true,
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			assert.NoError(afero.WriteFile(tc.fs, "/etc/passwd", []byte(tc.passwdContents), 0o755))
			if tc.readonly {
				tc.fs = afero.NewReadOnlyFs(tc.fs)
			}
			authorized := map[string]bool{}
			if tc.alreadyDeployed {
				authorized["user:ssh-rsa testkey"] = true
			}
			sshAccess := SSHAccess{
				fs: tc.fs,
				userManager: LinuxUserManager{
					fs:      tc.fs,
					passwd:  passwd.Passwd{},
					creator: tc.userCreator,
				},
				mux:        sync.Mutex{},
				authorized: authorized,
			}
			err := sshAccess.DeploySSHAuthorizedKey(context.Background(), authorizedKey)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			if tc.wantFile {
				fileContents, err := afero.ReadFile(tc.fs, "/home/user/.ssh/authorized_keys.d/debugd")
				assert.NoError(err)
				assert.Equal(tc.wantFileContents, string(fileContents))
			} else {
				exists, err := afero.Exists(tc.fs, "/home/user/.ssh/authorized_keys.d/debugd")
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
