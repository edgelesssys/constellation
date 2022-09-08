/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package proto

import (
	"context"
	"errors"
	"io"

	"github.com/edgelesssys/constellation/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
)

// RecoverClient wraps a recoverAPI client and the connection to it.
type RecoverClient struct {
	conn       *grpc.ClientConn
	recoverapi recoverproto.APIClient
}

// Connect connects the client to a given server, using the handed
// Validators for the attestation of the connection.
// The connection must be closed using Close(). If connect is
// called on a client that already has a connection, the old
// connection is closed.
func (c *RecoverClient) Connect(endpoint string, validators atls.Validator) error {
	creds := atlscredentials.New(nil, []atls.Validator{validators})

	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	_ = c.Close()
	c.conn = conn
	c.recoverapi = recoverproto.NewAPIClient(conn)
	return nil
}

// Close closes the grpc connection of the client.
// Close is idempotent and can be called on non connected clients
// without returning an error.
func (c *RecoverClient) Close() error {
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
func (c *RecoverClient) Recover(ctx context.Context, masterSecret, salt []byte) (retErr error) {
	if c.recoverapi == nil {
		return errors.New("client is not connected")
	}

	measurementSecret, err := attestation.DeriveMeasurementSecret(masterSecret, salt)
	if err != nil {
		return err
	}

	recoverclient, err := c.recoverapi.Recover(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := recoverclient.CloseSend(); err != nil {
			multierr.AppendInto(&retErr, err)
		}
	}()

	if err := recoverclient.Send(&recoverproto.RecoverMessage{
		Request: &recoverproto.RecoverMessage_MeasurementSecret{
			MeasurementSecret: measurementSecret,
		},
	}); err != nil {
		return err
	}
	res, err := recoverclient.Recv()
	if err != nil {
		return err
	}

	stateDiskKey, err := deriveStateDiskKey(masterSecret, salt, res.DiskUuid)
	if err != nil {
		return err
	}

	if err := recoverclient.Send(&recoverproto.RecoverMessage{
		Request: &recoverproto.RecoverMessage_StateDiskKey{
			StateDiskKey: stateDiskKey,
		},
	}); err != nil {
		return err
	}

	if _, err := recoverclient.Recv(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// deriveStateDiskKey derives a state disk key from a master key, a salt, and a disk UUID.
func deriveStateDiskKey(masterKey, salt []byte, diskUUID string) ([]byte, error) {
	return crypto.DeriveKey(masterKey, salt, []byte(crypto.HKDFInfoPrefix+diskUUID), crypto.StateDiskKeyLength)
}
