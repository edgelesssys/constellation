/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"os/exec"

	"github.com/coreos/go-systemd/v22/dbus"
)

// wraps go-systemd dbus.
type dbusWrapper struct{}

func (d *dbusWrapper) NewSystemConnectionContext(ctx context.Context) (dbusConn, error) {
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, err
	}
	return &dbusConnWrapper{
		conn: conn,
	}, nil
}

type dbusConnWrapper struct {
	conn *dbus.Conn
}

func (c *dbusConnWrapper) StartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	return c.conn.StartUnitContext(ctx, name, mode, ch)
}

func (c *dbusConnWrapper) StopUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	return c.conn.StopUnitContext(ctx, name, mode, ch)
}

func (c *dbusConnWrapper) ResetFailedUnitContext(ctx context.Context, name string) error {
	return c.conn.ResetFailedUnitContext(ctx, name)
}

func (c *dbusConnWrapper) RestartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	return c.conn.RestartUnitContext(ctx, name, mode, ch)
}

func (c *dbusConnWrapper) ReloadContext(ctx context.Context) error {
	return c.conn.ReloadContext(ctx)
}

func (c *dbusConnWrapper) Close() {
	c.conn.Close()
}

type journalctlWrapper struct{}

func (j *journalctlWrapper) readJournal(unit string) string {
	out, _ := exec.CommandContext(context.Background(), "journalctl", "-u", unit, "--no-pager").CombinedOutput()
	return string(out)
}

/*
// Preferably, we would use the systemd journal API directly.
// However, this requires linking against systemd libraries, so we go with the easier journalctl command for now.

type sdJournalWrapper struct{}

// readJournal reads the journal for a specific unit.
func (s *sdJournalWrapper) readJournal(unit string) string {
	journal, err := sdjournal.NewJournal()
	if err != nil {
		log.Printf("opening journal: %s", err)
		return ""
	}
	defer journal.Close()

	// Filter the journal for the specified unit
	filters := []string{
		fmt.Sprintf("_SYSTEMD_UNIT=%s", unit),
		fmt.Sprintf("UNIT=%s", unit),
		fmt.Sprintf("OBJECT_SYSTEMD_UNIT=%s", unit),
		fmt.Sprintf("_SYSTEMD_SLICE=%s", unit),
		fmt.Sprintf("_SYSTEMD_USER_UNIT=%s", unit),
		fmt.Sprintf("USER_UNIT=%s", unit),
		fmt.Sprintf("COREDUMP_USER_UNIT=%s", unit),
		fmt.Sprintf("OBJECT_SYSTEMD_USER_UNIT=%s", unit),
		fmt.Sprintf("_SYSTEMD_USER_SLICE=%s", unit),
	}
	for _, filter := range filters {
		if err := journal.AddMatch(filter); err != nil {
			log.Printf("applying filter %q: %s", filter, err)
			return ""
		}
		if err := journal.AddDisjunction(); err != nil {
			log.Printf("adding disjunction to journal filter: %s", err)
			return ""
		}
	}

	// Seek to the beginning of the journal
	if err := journal.SeekHead(); err != nil {
		log.Printf("seeking journal tail: %s", err)
		return ""
	}

	// Iterate over the journal entries
	var previousCursor string
	journalLog := &strings.Builder{}
	for {
		if _, err := journal.Next(); err != nil {
			log.Printf("getting next entry in journal: %s", err)
			return ""
		}

		entry, err := journal.GetEntry()
		if err != nil {
			log.Printf("getting journal entry: %s", err)
			return ""
		}

		// Abort if we reached the end of the journal, i.e. the cursor didn't change
		if entry.Cursor == previousCursor {
			break
		}
		previousCursor = entry.Cursor

		if _, err := journalLog.WriteString(entry.Fields[sdjournal.SD_JOURNAL_FIELD_MESSAGE] + "\n"); err != nil {
			log.Printf("copying journal entry to buffer: %s", err)
			return ""
		}
	}

	return strings.TrimSpace(journalLog.String())
}
*/
