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

// ErrUserOrGroupAlreadyExists is the Go error converted from the result of useradd or groupadd when an user or group already exists.
var ErrUserOrGroupAlreadyExists = errors.New("user or group already exists")

type passwdParser interface {
	Parse(fs afero.Fs) (Entries, error)
}

type userCreator interface {
	CreateUser(ctx context.Context, username string) error
	CreateUserWithSpecificUIDAndGID(ctx context.Context, username string, uid int, gid int) error
}

// LinuxUser holds relevant information about a linux user (subset of /etc/passwd).
type LinuxUser struct {
	Username string
	Home     string
	UID      int
	GID      int
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
	uids          []int
	createUserErr error
	currentUID    int
}

// CreateUser for StubUserCreator creates an user for an unit test environment.
func (s *StubUserCreator) CreateUser(ctx context.Context, username string) error {
	if stringInSlice(username, s.usernames) {
		// do not fail if user already exists
		return nil
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
		lineToWrite := fmt.Sprintf("%s:x:%d:%d:%s:/var/home/%s:/bin/bash\n", username, s.currentUID, s.currentUID, username, username)
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

	s.currentUID++
	s.usernames = append(s.usernames, username)

	return nil
}

// CreateUserWithSpecificUIDAndGID for StubUserCreator creates an user with a specific UID and GID for an unit test environment.
func (s *StubUserCreator) CreateUserWithSpecificUIDAndGID(ctx context.Context, username string, uid int, gid int) error {
	if stringInSlice(username, s.usernames) {
		// do not fail if user already exists
		return nil
	}
	if intInSlice(uid, s.uids) {
		return errors.New("uid is already used by another user")
	}

	if s.createUserErr != nil {
		return s.createUserErr
	}

	// If no predefined error is supposed to happen, increase the UID (unless the file system code fails)
	if s.fs != nil {
		lineToWrite := fmt.Sprintf("%s:x:%d:%d:%s:/var/home/%s:/bin/bash\n", username, uid, gid, username, username)
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

	// Mark UID as used (we don't track GIDs though, as multiple users can belong to one GID)
	s.uids = append(s.uids, uid)

	// Avoid potential collisions
	if s.currentUID == uid {
		s.currentUID++
	}

	s.usernames = append(s.usernames, username)

	return nil
}

// GetLinuxUser tries to find an existing linux user in /etc/passwd.
func (l *LinuxUserManager) GetLinuxUser(username string) (LinuxUser, error) {
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
		UID:      uid,
		GID:      gid,
	}, nil
}

// EnsureLinuxUserExists will try to create the user specified by username and call GetLinuxUser to retrieve user information.
func (l *LinuxUserManager) EnsureLinuxUserExists(ctx context.Context, username string) (LinuxUser, error) {
	// try to create user (even if it already exists)
	if err := l.Creator.CreateUser(ctx, username); err != nil && !errors.Is(err, ErrUserOrGroupAlreadyExists) {
		return LinuxUser{}, err
	}

	return l.GetLinuxUser(username)
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

// intInSlice checks if a given string exists in a slice of strings.
func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
