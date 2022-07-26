package proto

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"google.golang.org/grpc"
)

// KeyClient wraps a KeyAPI client and the connection to it.
type KeyClient struct {
	conn   *grpc.ClientConn
	keyapi keyproto.APIClient
}

// Connect connects the client to a given server, using the handed
// Validators for the attestation of the connection.
// The connection must be closed using Close(). If connect is
// called on a client that already has a connection, the old
// connection is closed.
func (c *KeyClient) Connect(endpoint string, validators []atls.Validator) error {
	creds := atlscredentials.New(nil, validators)

	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	_ = c.Close()
	c.conn = conn
	c.keyapi = keyproto.NewAPIClient(conn)
	return nil
}

// Close closes the grpc connection of the client.
// Close is idempotent and can be called on non connected clients
// without returning an error.
func (c *KeyClient) Close() error {
	if c.conn == nil {
		return nil
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	c.conn = nil
	return nil
}

// PushStateDiskKey pushes the state disk key to a constellation instance in recovery mode.
// The state disk key must be derived from the UUID of the state disk and the master key.
func (c *KeyClient) PushStateDiskKey(ctx context.Context, stateDiskKey, measurementSecret []byte) error {
	if c.keyapi == nil {
		return errors.New("client is not connected")
	}

	req := &keyproto.PushStateDiskKeyRequest{
		StateDiskKey:      stateDiskKey,
		MeasurementSecret: measurementSecret,
	}

	_, err := c.keyapi.PushStateDiskKey(ctx, req)
	return err
}
