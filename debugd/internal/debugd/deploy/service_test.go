/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/logger"
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
				log:                      logger.NewTest(t),
				dbus:                     &tc.dbus,
				journal:                  &stubJournalReader{},
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
					reloadErr: errors.New("reload error"),
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
				log:                      logger.NewTest(t),
				dbus:                     &tc.dbus,
				journal:                  &stubJournalReader{},
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

func TestOverrideServiceUnitExecStart(t *testing.T) {
	testCases := map[string]struct {
		dbus                stubDbus
		unitName, execStart string
		readonly            bool
		wantErr             bool
		wantFileContents    string
		wantActionCalls     []dbusConnActionInput
		wantReloads         int
	}{
		"override works": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			unitName:         "test",
			execStart:        "/run/state/bin/test",
			wantFileContents: "[Service]\nExecStart=\nExecStart=/run/state/bin/test\n",
			wantActionCalls: []dbusConnActionInput{
				{name: "test.service", mode: "replace"},
			},
			wantReloads: 1,
		},
		"unit name invalid": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			unitName:  "invalid name",
			execStart: "/run/state/bin/test",
			wantErr:   true,
		},
		"exec start invalid": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			unitName:  "test",
			execStart: "/run/state/bin/\r\ntest",
			wantErr:   true,
		},
		"write fails": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result: "done",
				},
			},
			unitName:  "test",
			execStart: "/run/state/bin/test",
			readonly:  true,
			wantErr:   true,
		},
		"reload fails but restart is still attempted": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result:    "done",
					reloadErr: errors.New("reload error"),
				},
			},
			unitName:         "test",
			execStart:        "/run/state/bin/test",
			wantFileContents: "[Service]\nExecStart=\nExecStart=/run/state/bin/test\n",
			wantActionCalls: []dbusConnActionInput{
				{name: "test.service", mode: "replace"},
			},
			wantReloads: 1,
		},
		"restart fails": {
			dbus: stubDbus{
				conn: &fakeDbusConn{
					result:    "done",
					actionErr: errors.New("action error"),
				},
			},
			unitName:  "test",
			execStart: "/run/state/bin/test",
			wantErr:   true,
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
				log:                      logger.NewTest(t),
				dbus:                     &tc.dbus,
				journal:                  &stubJournalReader{},
				fs:                       fs,
				systemdUnitFilewriteLock: sync.Mutex{},
			}
			err := manager.OverrideServiceUnitExecStart(context.Background(), tc.unitName, tc.execStart)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			fileContents, err := afero.ReadFile(fs, "/run/systemd/system/test.service.d/override.conf")
			assert.NoError(err)
			assert.Equal(tc.wantFileContents, string(fileContents))
			assert.Equal(tc.wantActionCalls, tc.dbus.conn.(*fakeDbusConn).inputs)
			assert.Equal(tc.wantReloads, tc.dbus.conn.(*fakeDbusConn).reloadCalls)
		})
	}
}

type stubDbus struct {
	conn    dbusConn
	connErr error
}

func (s *stubDbus) NewSystemConnectionContext(_ context.Context) (dbusConn, error) {
	return s.conn, s.connErr
}

type dbusConnActionInput struct {
	name string
	mode string
}

type fakeDbusConn struct {
	inputs      []dbusConnActionInput
	result      string
	reloadCalls int

	jobID     int
	actionErr error
	reloadErr error
}

func (c *fakeDbusConn) StartUnitContext(_ context.Context, name string, mode string, ch chan<- string) (int, error) {
	c.inputs = append(c.inputs, dbusConnActionInput{name: name, mode: mode})
	ch <- c.result

	return c.jobID, c.actionErr
}

func (c *fakeDbusConn) StopUnitContext(_ context.Context, name string, mode string, ch chan<- string) (int, error) {
	c.inputs = append(c.inputs, dbusConnActionInput{name: name, mode: mode})
	ch <- c.result

	return c.jobID, c.actionErr
}

func (c *fakeDbusConn) ResetFailedUnitContext(_ context.Context, _ string) error {
	return nil
}

func (c *fakeDbusConn) RestartUnitContext(_ context.Context, name string, mode string, ch chan<- string) (int, error) {
	c.inputs = append(c.inputs, dbusConnActionInput{name: name, mode: mode})
	ch <- c.result

	return c.jobID, c.actionErr
}

func (c *fakeDbusConn) ReloadContext(_ context.Context) error {
	c.reloadCalls++

	return c.reloadErr
}

func (c *fakeDbusConn) Close() {}

type stubJournalReader struct{}

func (s *stubJournalReader) readJournal(_ string) string {
	return ""
}
