package deploy

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
)

// wraps go-systemd dbus
type dbusWrapper struct{}

func (d *dbusWrapper) NewSystemdConnectionContext(ctx context.Context) (dbusConn, error) {
	conn, err := dbus.NewSystemdConnectionContext(ctx)
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
func (c *dbusConnWrapper) RestartUnitContext(ctx context.Context, name string, mode string, ch chan<- string) (int, error) {
	return c.conn.RestartUnitContext(ctx, name, mode, ch)
}
func (c *dbusConnWrapper) ReloadContext(ctx context.Context) error {
	return c.conn.ReloadContext(ctx)
}
