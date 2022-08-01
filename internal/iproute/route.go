package iproute

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// AddToLocalRoutingTable adds the IP to the local routing table.
func AddToLocalRoutingTable(ctx context.Context, ip string) error {
	return manipulateLocalRoutingTable(ctx, "add", ip)
}

// RemoveFromLocalRoutingTable removes the IPfrom the local routing table.
func RemoveFromLocalRoutingTable(ctx context.Context, ip string) error {
	return manipulateLocalRoutingTable(ctx, "del", ip)
}

func manipulateLocalRoutingTable(ctx context.Context, action string, ip string) error {
	// https://github.com/GoogleCloudPlatform/guest-agent/blob/792fce795218633bcbde505fb3457a0b24f26d37/google_guest_agent/addresses.go#L179
	if !strings.Contains(ip, "/") {
		ip = ip + "/32"
	}

	args := []string{"route", action, "to", "local", ip, "scope", "host", "dev", "ens3", "proto", "66"}
	_, err := exec.CommandContext(ctx, "ip", args...).Output()
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return fmt.Errorf("ip route %s: %w", action, err)
	}
	if exitErr.ExitCode() == 2 {
		// "RTNETLINK answers: File exists" or "RTNETLINK answers: No such process"
		//
		// Ignore, expected in case of adding an existing route or deleting a route
		// that does not exist.
		return nil
	}
	return fmt.Errorf("ip route %s (code %v) with: %s", action, exitErr.ExitCode(), exitErr.Stderr)
}
