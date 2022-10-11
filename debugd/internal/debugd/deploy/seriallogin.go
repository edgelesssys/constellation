/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	"github.com/spf13/afero"
)

// EnableAutoLogin installs a systemd unit override that allows passwordless root login
// on the serial console.
func EnableAutoLogin(ctx context.Context, fs afero.Fs, serviceManager serviceManager) error {
	if err := fs.MkdirAll(path.Dir(debugd.GettyAutologinOverrideFilename), os.ModePerm); err != nil {
		return fmt.Errorf("creating getty autologin override directory: %w", err)
	}
	if err := afero.WriteFile(fs, debugd.GettyAutologinOverrideFilename,
		[]byte(debugd.GettyAutologinOverrideUnitContents), os.ModePerm); err != nil {
		return fmt.Errorf("writing getty autologin override unit: %w", err)
	}
	if err := serviceManager.SystemdAction(ctx, ServiceManagerRequest{
		Action: Reload,
	}); err != nil {
		return fmt.Errorf("reloading systemd units: %w", err)
	}
	if err := serviceManager.SystemdAction(ctx, ServiceManagerRequest{
		Action: Restart,
		Unit:   "serial-getty@ttyS0.service",
	}); err != nil {
		return fmt.Errorf("restarting getty: %w", err)
	}
	return nil
}
