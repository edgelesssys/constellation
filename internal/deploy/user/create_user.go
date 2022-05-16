package user

import (
	"context"
	"fmt"
	"os/exec"
)

type Unix struct{}

// reference: https://man7.org/linux/man-pages/man8/useradd.8.html#EXIT_VALUES
const exitCodeUsernameAlreadyInUse = 9

// CreateUser creates a new user with sudo access. Returns successfully if creation succeeds or user existed already.
func (u Unix) CreateUser(ctx context.Context, username string) error {
	cmd := exec.CommandContext(ctx, "useradd", "-m", "-G", "wheel,sudo", username)
	if err := cmd.Run(); err != nil {
		// do not fail if user already exists
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == exitCodeUsernameAlreadyInUse {
			return nil
		}
		return fmt.Errorf("creating a new user failed: %w", err)
	}
	return nil
}
