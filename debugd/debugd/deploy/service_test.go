package deploy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemdAction(t *testing.T) {
	unitName := "example.service"

	testCases := map[string]struct {
		dbus    stubDbus
		action  SystemdAction
		wantErr bool
	}{
		"start works": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			action:  Start,
			wantErr: false,
		},
		"stop works": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			action:  Stop,
			wantErr: false,
		},
		"restart works": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			action:  Restart,
			wantErr: false,
		},
		"reload works": {
			dbus: stubDbus{
				conn: &fakeDbusConn{},
			},
			action:  Reload,
			wantErr: false,
		},
		"unknown action": {
			dbus: stubDbus{
				conn: &fakeDbusConn{},
			},
			action:  Unknown,
			wantErr: true,
		},
		"action fails": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					actionErr: errors.New("action fails"),
				},
			},
			action:  Start,
			wantErr: true,
		},
		"action result is failure": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "failure",
				},
			},
			action:  Start,
			wantErr: true,
		},
		"newConn fails": {
			dbus: stubDbus{
				connErr: errors.New("newConn fails"),
			},
			action:  Start,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			manager := ServiceManager{
				dbus:                     &tc.dbus,
				fs:                       fs,
				systemdUnitFilewriteLock: sync.Mutex{},
			}
			err := manager.SystemdAction(context.Background(), ServiceManagerRequest{
				Unit:   unitName,
				Action: tc.action,
			})

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestWriteSystemdUnitFile(t *testing.T) {
	testCases := map[string]struct {
		dbus             stubDbus
		unit             SystemdUnit
		readonly         bool
		wantErr          bool
		wantFileContents string
	}{
		"start works": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			unit: SystemdUnit{
				Name:     "test.service",
				Contents: "testservicefilecontents",
			},
			wantErr:          false,
			wantFileContents: "testservicefilecontents",
		},
		"write fails": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			unit: SystemdUnit{
				Name:     "test.service",
				Contents: "testservicefilecontents",
			},
			readonly: true,
			wantErr:  true,
		},
		"systemd reload fails": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					actionErr: errors.New("reload error"),
				},
			},
			unit: SystemdUnit{
				Name:     "test.service",
				Contents: "testservicefilecontents",
			},
			readonly: false,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			assert.NoError(fs.MkdirAll(systemdUnitFolder, 0o755))
			if tc.readonly {
				fs = afero.NewReadOnlyFs(fs)
			}
			manager := ServiceManager{
				dbus:                     &tc.dbus,
				fs:                       fs,
				systemdUnitFilewriteLock: sync.Mutex{},
			}
			err := manager.WriteSystemdUnitFile(context.Background(), tc.unit)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			fileContents, err := afero.ReadFile(fs, fmt.Sprintf("%s/%s", systemdUnitFolder, tc.unit.Name))
			assert.NoError(err)
			assert.Equal(tc.wantFileContents, string(fileContents))
		})
	}
}

type stubDbus struct {
	conn    dbusConn
	connErr error
}

func (s *stubDbus) NewSystemdConnectionContext(ctx context.Context) (dbusConn, error) {
	return s.conn, s.connErr
}

type dbusConnActionInput struct {
	name string
	mode string
}

type fakeDbusConn struct {
	inputs []dbusConnActionInput
	result string

	jobID     int
	actionErr error
}

func (f *fakeDbusConn) StartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	f.inputs = append(f.inputs, dbusConnActionInput{name: name, mode: mode})
	go func() {
		ch <- f.result
	}()

	return f.jobID, f.actionErr
}

func (f *fakeDbusConn) StopUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	f.inputs = append(f.inputs, dbusConnActionInput{name: name, mode: mode})
	go func() {
		ch <- f.result
	}()
	return f.jobID, f.actionErr
}

func (f *fakeDbusConn) RestartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	f.inputs = append(f.inputs, dbusConnActionInput{name: name, mode: mode})
	go func() {
		ch <- f.result
	}()
	return f.jobID, f.actionErr
}

func (s *fakeDbusConn) ReloadContext(ctx context.Context) error {
	return s.actionErr
}
