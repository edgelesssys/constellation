package proto

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	tlsConfig, err := atls.CreateAttestationClientTLSConfig(nil, validators)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
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
func (c *KeyClient) PushStateDiskKey(ctx context.Context, stateDiskKey []byte) error {
	if c.keyapi == nil {
		return errors.New("client is not connected")
	}

	req := &keyproto.PushStateDiskKeyRequest{
		StateDiskKey: stateDiskKey,
	}

	_, err := c.keyapi.PushStateDiskKey(ctx, req)
	return err
}
