package k8sapi

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

func RestartSystemdUnit(unit string) error {
	ctx := context.Background()
	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return fmt.Errorf("establishing systemd connection failed: %w", err)
	}

	restartChan := make(chan string)
	if _, err := conn.RestartUnitContext(ctx, unit, "replace", restartChan); err != nil {
		return fmt.Errorf("restarting systemd unit \"%v\" failed: %w", unit, err)
	}

	// Wait for the restart to finish and actually check if it was
	// successful or not.
	result := <-restartChan

	switch result {
	case "done":
		return nil

	default:
		return fmt.Errorf("restarting systemd unit \"%v\" failed: expected %v but received %v", unit, "done", result)
	}
}
