package user

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

// Unix defines an user creation interface for UNIX systems.
type Unix struct{}

// reference: https://man7.org/linux/man-pages/man8/useradd.8.html#EXIT_VALUES
const exitCodeAlreadyInUse = 9

// CreateUser creates a new user with sudo access.
func (u Unix) CreateUser(ctx context.Context, username string) error {
	cmd := exec.CommandContext(ctx, "useradd", "-m", "-G", "wheel,sudo", username)
	if err := cmd.Run(); err != nil {
		// do not fail if user already exists
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == exitCodeAlreadyInUse {
			return ErrUserOrGroupAlreadyExists
		}
		return fmt.Errorf("creating a new user failed: %w", err)
	}
	return nil
}

// CreateUserWithSpecificUIDAndGID creates a new user with sudo access and a specific UID and GID.
func (u Unix) CreateUserWithSpecificUIDAndGID(ctx context.Context, username string, uid int, gid int) error {
	// Add group first with the targeted gid
	cmd := exec.CommandContext(ctx, "groupadd", "-g", strconv.Itoa(gid), username)
	if err := cmd.Run(); err != nil {
		// do not fail if group already exists
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == exitCodeAlreadyInUse {
			return ErrUserOrGroupAlreadyExists
		}
		return fmt.Errorf("creating a new group failed: %w", err)
	}

	// Then, create the user with both the UID and GID assigned.
	cmd = exec.CommandContext(ctx, "useradd", "-m", "-G", "wheel,sudo", "-u", strconv.Itoa(uid), "-g", strconv.Itoa(gid), username)
	if err := cmd.Run(); err != nil {
		// do not fail if user already exists
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == exitCodeAlreadyInUse {
			return ErrUserOrGroupAlreadyExists
		}
		return fmt.Errorf("creating a new user failed: %w", err)
	}
	return nil
}
