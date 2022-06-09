package deploy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/edgelesssys/constellation/debugd/debugd"
	"github.com/spf13/afero"
)

const (
	systemdUnitFolder = "/etc/systemd/system"
)

//go:generate stringer -type=SystemdAction
type SystemdAction uint32

const (
	Unknown SystemdAction = iota
	Start
	Stop
	Restart
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
	dbus                     dbusClient
	fs                       afero.Fs
	systemdUnitFilewriteLock sync.Mutex
}

// NewServiceManager creates a new ServiceManager.
func NewServiceManager() *ServiceManager {
	fs := afero.NewOsFs()
	return &ServiceManager{
		dbus:                     &dbusWrapper{},
		fs:                       fs,
		systemdUnitFilewriteLock: sync.Mutex{},
	}
}

type dbusClient interface {
	// NewSystemConnectionContext establishes a connection to the system bus and authenticates.
	// Callers should call Close() when done with the connection.
	NewSystemdConnectionContext(ctx context.Context) (dbusConn, error)
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
}

// SystemdAction will perform a systemd action on a service unit (start, stop, restart, reload).
func (s *ServiceManager) SystemdAction(ctx context.Context, request ServiceManagerRequest) error {
	conn, err := s.dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return fmt.Errorf("establishing systemd connection: %w", err)
	}

	resultChan := make(chan string)
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
		return errors.New("unknown systemd action: " + request.Action.String())
	}
	if err != nil {
		return fmt.Errorf("performing systemd action %v on unit %v: %w", request.Action, request.Unit, err)
	}

	if request.Action == Reload {
		log.Println("daemon-reload succeeded")
		return nil
	}
	// Wait for the action to finish and then check if it was
	// successful or not.
	result := <-resultChan

	switch result {
	case "done":
		log.Printf("%s on systemd unit %s succeeded\n", request.Action, request.Unit)
		return nil

	default:
		return fmt.Errorf("performing action %v on systemd unit \"%v\" failed: expected \"%v\" but received \"%v\"", request.Action, request.Unit, "done", result)
	}
}

// WriteSystemdUnitFile will write a systemd unit to disk.
func (s *ServiceManager) WriteSystemdUnitFile(ctx context.Context, unit SystemdUnit) error {
	log.Printf("Writing systemd unit file: %s/%s\n", systemdUnitFolder, unit.Name)
	s.systemdUnitFilewriteLock.Lock()
	defer s.systemdUnitFilewriteLock.Unlock()
	if err := afero.WriteFile(s.fs, fmt.Sprintf("%s/%s", systemdUnitFolder, unit.Name), []byte(unit.Contents), 0o644); err != nil {
		return fmt.Errorf("writing systemd unit file \"%v\": %w", unit.Name, err)
	}

	if err := s.SystemdAction(ctx, ServiceManagerRequest{Unit: unit.Name, Action: Reload}); err != nil {
		return fmt.Errorf("performing systemd daemon-reload: %w", err)
	}

	log.Printf("Wrote systemd unit file: %s/%s and performed daemon-reload\n", systemdUnitFolder, unit.Name)

	return nil
}

// DeployDefaultServiceUnit will write the default "coordinator.service" unit file.
func DeployDefaultServiceUnit(ctx context.Context, serviceManager *ServiceManager) error {
	if err := serviceManager.WriteSystemdUnitFile(ctx, SystemdUnit{
		Name:     debugd.CoordinatorSystemdUnitName,
		Contents: debugd.CoordinatorSystemdUnitContents,
	}); err != nil {
		return fmt.Errorf("writing systemd unit file %q: %w", debugd.CoordinatorSystemdUnitName, err)
	}

	// try to start the default service if the binary exists but ignore failure.
	// this is meant to start the coordinator after a reboot
	// if a coordinator binary was uploaded before.
	if ok, err := afero.Exists(serviceManager.fs, debugd.CoordinatorDeployFilename); ok && err == nil {
		_ = serviceManager.SystemdAction(ctx, ServiceManagerRequest{
			Unit:   debugd.CoordinatorSystemdUnitName,
			Action: Start,
		})
	}

	return nil
}
