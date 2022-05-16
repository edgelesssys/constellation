package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/afero"
)

// ErrUserDoesNotExist is returned by GetLinuxUser if a linux user does not exist yet.
var ErrUserDoesNotExist = errors.New("user does not exist")

type passwdParser interface {
	Parse(fs afero.Fs) (Entries, error)
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
	Fs      afero.Fs
	Passwd  passwdParser
	Creator userCreator
}

// NewLinuxUserManager creates a new LinuxUserManager.
func NewLinuxUserManager(fs afero.Fs) LinuxUserManager {
	return LinuxUserManager{
		Fs:      fs,
		Passwd:  Passwd{},
		Creator: Unix{},
	}
}

// NewLinuxUserManagerFake creates a new LinuxUserManager that is used for unit tests.
func NewLinuxUserManagerFake(fs afero.Fs) LinuxUserManager {
	return LinuxUserManager{
		Fs:      fs,
		Passwd:  Passwd{},
		Creator: &StubUserCreator{fs: fs},
	}
}

// StubUserCreator is used for unit tests.
type StubUserCreator struct {
	fs            afero.Fs
	usernames     []string
	createUserErr error
	currentUID    int
}

func (s *StubUserCreator) CreateUser(ctx context.Context, username string) error {
	if stringInSlice(username, s.usernames) {
		return errors.New("username already exists")
	}

	// We want created users to start at UID 1000
	if s.currentUID == 0 {
		s.currentUID = 1000
	}

	if s.createUserErr != nil {
		return s.createUserErr
	}

	// If no predefined error is supposed to happen, increase the UID (unless the file system code fails)
	if s.fs != nil {
		lineToWrite := fmt.Sprintf("%s:x:%d:%d:%s:/home/%s:/bin/bash\n", username, s.currentUID, s.currentUID, username, username)
		file, err := s.fs.OpenFile("/etc/passwd", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()

		n, err := file.WriteString(lineToWrite)

		if err != nil {
			return err
		} else if n != len(lineToWrite) {
			return errors.New("written text too short")
		}
	}

	s.currentUID += 1
	s.usernames = append(s.usernames, username)

	return nil
}

// getLinuxUser tries to find an existing linux user in /etc/passwd.
func (l *LinuxUserManager) getLinuxUser(username string) (LinuxUser, error) {
	entries, err := l.Passwd.Parse(l.Fs)
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
	if err := l.Creator.CreateUser(ctx, username); err != nil {
		return LinuxUser{}, err
	}

	return l.getLinuxUser(username)
}

// stringInSlice checks if a given string exists in a slice of strings.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
