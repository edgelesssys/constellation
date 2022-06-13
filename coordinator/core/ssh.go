package core

import (
	"context"

	"github.com/edgelesssys/constellation/internal/deploy/ssh"
)

// CreateSSHUsers creates UNIX users with respective SSH access on the system the coordinator is running on when defined in the config.
func (c *Core) CreateSSHUsers(sshUserKeys []ssh.UserKey) error {
	sshAccess := ssh.NewAccess(c.linuxUserManager)
	ctx := context.Background()

	for _, pair := range sshUserKeys {
		if err := sshAccess.DeployAuthorizedKey(ctx, pair); err != nil {
			return err
		}
	}

	return nil
}
