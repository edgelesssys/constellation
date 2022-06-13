package ssh

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/edgelesssys/constellation/internal/deploy/user"
)

// UserKey describes an user that should be created with a corresponding public SSH key.
type UserKey struct {
	Username  string
	PublicKey string
}

// Access reads SSH public keys from a channel, creates the specified users if required and writes the public keys to the users authorized_keys file.
type Access struct {
	userManager user.LinuxUserManager
	authorized  map[string]bool
	mux         sync.Mutex
}

// NewAccess creates a new Access.
func NewAccess(userManager user.LinuxUserManager) *Access {
	return &Access{
		userManager: userManager,
		mux:         sync.Mutex{},
		authorized:  map[string]bool{},
	}
}

// alreadyAuthorized checks if key was written to authorized keys before.
func (s *Access) alreadyAuthorized(sshKey UserKey) bool {
	_, ok := s.authorized[fmt.Sprintf("%s:%s", sshKey.Username, sshKey.PublicKey)]
	return ok
}

// rememberAuthorized marks this key as already written to authorized keys..
func (s *Access) rememberAuthorized(sshKey UserKey) {
	s.authorized[fmt.Sprintf("%s:%s", sshKey.Username, sshKey.PublicKey)] = true
}

// DeployAuthorizedKey takes an user & public key pair, creates the user if required and deploy a SSH key for them.
func (s *Access) DeployAuthorizedKey(ctx context.Context, sshKey UserKey) error {
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
	authorizedKeysD := fmt.Sprintf("%s/authorized_keys.d", sshFolder)
	if err := s.userManager.Fs.MkdirAll(authorizedKeysD, 0o700); err != nil {
		return err
	}
	if err := s.userManager.Fs.Chown(sshFolder, user.UID, user.GID); err != nil {
		return err
	}
	if err := s.userManager.Fs.Chown(authorizedKeysD, user.UID, user.GID); err != nil {
		return err
	}
	authorizedKeysPath := fmt.Sprintf("%s/constellation-ssh-keys", authorizedKeysD)
	authorizedKeysFile, err := s.userManager.Fs.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = authorizedKeysFile.WriteString(fmt.Sprintf("%s\n", sshKey.PublicKey))
	if err != nil {
		return err
	}
	if err := authorizedKeysFile.Close(); err != nil {
		return err
	}
	if err := s.userManager.Fs.Chown(authorizedKeysPath, user.UID, user.GID); err != nil {
		return err
	}
	if err := s.userManager.Fs.Chmod(authorizedKeysPath, 0o644); err != nil {
		return err
	}
	s.rememberAuthorized(sshKey)
	log.Printf("Successfully authorized %s\n", sshKey.Username)
	return nil
}
