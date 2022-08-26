package deploy

import (
	"context"
	"fmt"
	"os/exec"
)

// DeleteUserPassword sets the user's password to an empty string
// effectively allowing anyone with access to the serial console to log in.
func DeleteUserPassword(ctx context.Context, user string) error {
	cmd := exec.CommandContext(ctx, "passwd", "-d", user)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deleting user password: %q %w", output, err)
	}
	return nil
}
