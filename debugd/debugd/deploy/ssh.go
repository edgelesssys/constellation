package deploy

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/edgelesssys/constellation/debugd/debugd/deploy/createuser"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy/passwd"
	"github.com/edgelesssys/constellation/debugd/ssh"
	"github.com/spf13/afero"
)

// SSHAccess reads ssh public keys from a channel, creates the specified users if required and writes the public keys to the users authorized_keys file.
type SSHAccess struct {
	fs          afero.Fs
	userManager LinuxUserManager
	authorized  map[string]bool
	mux         sync.Mutex
}

// NewSSHAccess creates a new SSHAccess.
func NewSSHAccess(fs afero.Fs) *SSHAccess {
	return &SSHAccess{
		fs: fs,
		userManager: LinuxUserManager{
			fs:      fs,
			passwd:  passwd.Passwd{},
			creator: createuser.Unix{},
		},
		mux:        sync.Mutex{},
		authorized: map[string]bool{},
	}
}

// alreadyAuthorized checks if key was written to authorized keys before.
func (s *SSHAccess) alreadyAuthorized(sshKey ssh.SSHKey) bool {
	_, ok := s.authorized[fmt.Sprintf("%s:%s", sshKey.Username, sshKey.KeyValue)]
	return ok
}

// rememberAuthorized marks this key as already written to authorized keys..
func (s *SSHAccess) rememberAuthorized(sshKey ssh.SSHKey) {
	s.authorized[fmt.Sprintf("%s:%s", sshKey.Username, sshKey.KeyValue)] = true
}

func (s *SSHAccess) DeploySSHAuthorizedKey(ctx context.Context, sshKey ssh.SSHKey) error {
	// allow only one thread to write to authorized keys, create users and update the authorized map at a time
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.alreadyAuthorized(sshKey) {
		return nil
	}
	log.Printf("Trying to deploy ssh key for %s\n", sshKey.Username)
	user, err := s.userManager.EnsureLinuxUserExists(ctx, sshKey.Username)
	if err != nil {
		return err
	}
	// CoreOS uses https://github.com/coreos/ssh-key-dir to search for ssh keys in ~/.ssh/authorized_keys.d/*
	sshFolder := fmt.Sprintf("%s/.ssh", user.Home)
	authorized_keys_d := fmt.Sprintf("%s/authorized_keys.d", sshFolder)
	if err := s.fs.MkdirAll(authorized_keys_d, 0o700); err != nil {
		return err
	}
	if err := s.fs.Chown(sshFolder, user.Uid, user.Gid); err != nil {
		return err
	}
	if err := s.fs.Chown(authorized_keys_d, user.Uid, user.Gid); err != nil {
		return err
	}
	authorizedKeysPath := fmt.Sprintf("%s/debugd", authorized_keys_d)
	authorizedKeysFile, err := s.fs.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = authorizedKeysFile.WriteString(fmt.Sprintf("%s %s\n", sshKey.KeyValue, sshKey.Username))
	if err != nil {
		return err
	}
	if err := authorizedKeysFile.Close(); err != nil {
		return err
	}
	if err := s.fs.Chown(authorizedKeysPath, user.Uid, user.Gid); err != nil {
		return err
	}
	if err := s.fs.Chmod(authorizedKeysPath, 0o644); err != nil {
		return err
	}
	s.rememberAuthorized(sshKey)
	log.Printf("Successfully authorized %s\n", sshKey.Username)
	return nil
}
