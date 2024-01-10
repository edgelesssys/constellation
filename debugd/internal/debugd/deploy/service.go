/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/afero"
)

const (
	systemdUnitFolder = "/run/systemd/system"
)

// systemdUnitNameRegexp is a regular expression that matches valid systemd unit names.
// This is only the unit name, without the .service suffix.
var systemdUnitNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9@._\-\\]+$`)

// SystemdAction encodes the available actions.
//
//go:generate stringer -type=SystemdAction
type SystemdAction uint32

const (
	// Unknown is the default SystemdAction and does nothing.
	Unknown SystemdAction = iota
	// Start a systemd service.
	Start
	// Stop a systemd service.
	Stop
	// Restart a systemd service.
	Restart
	// Reload a systemd service.
	Reload
)

// ServiceManagerRequest describes a requested ServiceManagerAction to be performed on a specified service unit.
type ServiceManagerRequest struct {
	Unit   string
	Action SystemdAction
}

// SystemdUnit describes a systemd service file including the unit name and contents.
type SystemdUnit struct {
	Name     string `yaml:"name"`
	Contents string `yaml:"contents"`
}

// ServiceManager receives ServiceManagerRequests and units via channels and performs the requests / creates the unit files.
type ServiceManager struct {
	log                      *slog.Logger
	dbus                     dbusClient
	fs                       afero.Fs
	systemdUnitFilewriteLock sync.Mutex
}

// NewServiceManager creates a new ServiceManager.
func NewServiceManager(log *slog.Logger) *ServiceManager {
	fs := afero.NewOsFs()
	return &ServiceManager{
		log:                      log,
		dbus:                     &dbusWrapper{},
		fs:                       fs,
		systemdUnitFilewriteLock: sync.Mutex{},
	}
}

type dbusClient interface {
	// NewSystemConnectionContext establishes a connection to the system bus and authenticates.
	// Callers should call Close() when done with the connection.
	NewSystemConnectionContext(ctx context.Context) (dbusConn, error)
}

type dbusConn interface {
	// StartUnitContext enqueues a start job and depending jobs, if any (unless otherwise
	// specified by the mode string).
	StartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error)
	// StopUnitContext is similar to StartUnitContext, but stops the specified unit
	// rather than starting it.
	StopUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error)
	// RestartUnitContext restarts a service. If a service is restarted that isn't
	// running it will be started.
	RestartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error)
	// ReloadContext instructs systemd to scan for and reload unit files. This is
	// an equivalent to systemctl daemon-reload.
	ReloadContext(ctx context.Context) error
	// Close closes the connection.
	Close()
}

// SystemdAction will perform a systemd action on a service unit (start, stop, restart, reload).
func (s *ServiceManager) SystemdAction(ctx context.Context, request ServiceManagerRequest) error {
	log := s.log.With(slog.String("unit", request.Unit), slog.String("action", request.Action.String()))
	conn, err := s.dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return fmt.Errorf("establishing systemd connection: %w", err)
	}
	defer conn.Close()

	resultChan := make(chan string, 1)
	switch request.Action {
	case Start:
		_, err = conn.StartUnitContext(ctx, request.Unit, "replace", resultChan)
	case Stop:
		_, err = conn.StopUnitContext(ctx, request.Unit, "replace", resultChan)
	case Restart:
		_, err = conn.RestartUnitContext(ctx, request.Unit, "replace", resultChan)
	case Reload:
		err = conn.ReloadContext(ctx)
	default:
		return fmt.Errorf("unknown systemd action: %s", request.Action.String())
	}
	if err != nil {
		return fmt.Errorf("performing systemd action %v on unit %v: %w", request.Action, request.Unit, err)
	}

	if request.Action == Reload {
		log.Info("daemon-reload succeeded")
		return nil
	}
	// Wait for the action to finish and then check if it was
	// successful or not.
	result := <-resultChan

	switch result {
	case "done":
		log.Info(fmt.Sprintf("%s on systemd unit %s succeeded", request.Action, request.Unit))
		return nil

	default:
		return fmt.Errorf("performing action %q on systemd unit %q failed: expected %q but received %q", request.Action.String(), request.Unit, "done", result)
	}
}

// WriteSystemdUnitFile will write a systemd unit to disk.
func (s *ServiceManager) WriteSystemdUnitFile(ctx context.Context, unit SystemdUnit) error {
	log := s.log.With(slog.String("unitFile", fmt.Sprintf("%s/%s", systemdUnitFolder, unit.Name)))
	log.Info("Writing systemd unit file")
	s.systemdUnitFilewriteLock.Lock()
	defer s.systemdUnitFilewriteLock.Unlock()
	if err := afero.WriteFile(s.fs, fmt.Sprintf("%s/%s", systemdUnitFolder, unit.Name), []byte(unit.Contents), 0o644); err != nil {
		return fmt.Errorf("writing systemd unit file \"%v\": %w", unit.Name, err)
	}

	if err := s.SystemdAction(ctx, ServiceManagerRequest{Unit: unit.Name, Action: Reload}); err != nil {
		return fmt.Errorf("performing systemd daemon-reload: %w", err)
	}

	log.Info("Wrote systemd unit file and performed daemon-reload")
	return nil
}

// OverrideServiceUnitExecStart will override the ExecStart of a systemd unit.
func (s *ServiceManager) OverrideServiceUnitExecStart(ctx context.Context, unitName, execStart string) error {
	log := s.log.With(slog.String("unitFile", fmt.Sprintf("%s/%s", systemdUnitFolder, unitName)))
	log.Info("Overriding systemd unit file execStart")
	if !systemdUnitNameRegexp.MatchString(unitName) {
		return fmt.Errorf("unit name %q is invalid", unitName)
	}
	// validate execStart (no newlines)
	if strings.Contains(execStart, "\n") || strings.Contains(execStart, "\r") {
		return fmt.Errorf("execStart must not contain newlines")
	}
	overrideUnitContents := fmt.Sprintf("[Service]\nExecStart=\nExecStart=%s\n", execStart)
	s.systemdUnitFilewriteLock.Lock()
	defer s.systemdUnitFilewriteLock.Unlock()
	path := filepath.Join(systemdUnitFolder, unitName+".service.d", "override.conf")
	if err := s.fs.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("creating systemd unit file override directory %q: %w", filepath.Dir(path), err)
	}
	if err := afero.WriteFile(s.fs, path, []byte(overrideUnitContents), 0o644); err != nil {
		return fmt.Errorf("writing systemd unit override file %q: %w", unitName, err)
	}
	if err := s.SystemdAction(ctx, ServiceManagerRequest{Unit: unitName, Action: Reload}); err != nil {
		// do not return early here
		// the "daemon-reload" command may return an unrelated error
		// and there is no way to know if the override was successful
		log.Warn("Failed to perform systemd daemon-reload: %v", err)
	}
	if err := s.SystemdAction(ctx, ServiceManagerRequest{Unit: unitName + ".service", Action: Restart}); err != nil {
		log.Warn("Failed to perform unit restart: %v", err)
		return fmt.Errorf("performing systemd unit restart: %w", err)
	}

	log.Info(fmt.Sprintf("Overrode systemd unit file execStart, performed daemon-reload and restarted unit %v", unitName))
	return nil
}
