package deploy

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/edgelesssys/constellation/debugd/debugd/deploy/createuser"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy/passwd"

	"github.com/spf13/afero"
)

// ErrUserDoesNotExist is returned by GetLinuxUser if a linux user does not exist yet.
var ErrUserDoesNotExist = errors.New("user does not exist")

type passwdParser interface {
	Parse(fs afero.Fs) (passwd.Entries, error)
}

type userCreator interface {
	CreateUser(ctx context.Context, username string) error
}

// LinuxUser holds relevant information about a linux user (subset of /etc/passwd).
type LinuxUser struct {
	Username string
	Home     string
	Uid      int
	Gid      int
}

// LinuxUserManager can retrieve information on linux users and create new users.
type LinuxUserManager struct {
	fs      afero.Fs
	passwd  passwdParser
	creator userCreator
}

// NewLinuxUserManager creates a new LinuxUserManager.
func NewLinuxUserManager(fs afero.Fs) *LinuxUserManager {
	return &LinuxUserManager{
		fs:      fs,
		passwd:  passwd.Passwd{},
		creator: createuser.Unix{},
	}
}

// getLinuxUser tries to find an existing linux user in /etc/passwd.
func (l *LinuxUserManager) getLinuxUser(username string) (LinuxUser, error) {
	entries, err := l.passwd.Parse(l.fs)
	if err != nil {
		return LinuxUser{}, err
	}
	if _, ok := entries[username]; !ok {
		return LinuxUser{}, ErrUserDoesNotExist
	}
	entry := entries[username]
	uid, err := strconv.Atoi(entry.Uid)
	if err != nil {
		return LinuxUser{}, fmt.Errorf("failed to parse users uid: %w", err)
	}
	gid, err := strconv.Atoi(entry.Gid)
	if err != nil {
		return LinuxUser{}, fmt.Errorf("failed to parse users gid: %w", err)
	}
	return LinuxUser{
		Username: username,
		Home:     entry.Home,
		Uid:      uid,
		Gid:      gid,
	}, nil

}

// EnsureLinuxUserExists will try to create the user specified by username and call GetLinuxUser to retrieve user information.
func (l *LinuxUserManager) EnsureLinuxUserExists(ctx context.Context, username string) (LinuxUser, error) {
	// try to create user (even if it already exists)
	if err := l.creator.CreateUser(ctx, username); err != nil {
		return LinuxUser{}, err
	}

	return l.getLinuxUser(username)
}
