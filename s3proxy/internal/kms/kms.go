/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package kms is used to interact with the Constellation keyservice.
So far it is a copy of the joinservice's kms package.
*/
package kms

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/edgelesssys/constellation/v2/keyservice/keyserviceproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client interacts with Constellation's keyservice.
type Client struct {
	log      *slog.Logger
	endpoint string
	grpc     grpcClient
}

// New creates a new KMS.
func New(log *slog.Logger, endpoint string) Client {
	return Client{
		log:      log,
		endpoint: endpoint,
		grpc:     client{},
	}
}

// GetDataKey returns a data encryption key for the given UUID.
func (c Client) GetDataKey(ctx context.Context, keyID string, length int) ([]byte, error) {
	log := c.log.With("keyID", keyID, "endpoint", c.endpoint)
	// the KMS does not use aTLS since traffic is only routed through the Constellation cluster
	// cluster internal connections are considered trustworthy
	log.Info("Connecting to KMS")
	conn, err := grpc.DialContext(ctx, c.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	log.Info("Requesting data key")
	res, err := c.grpc.GetDataKey(
		ctx,
		&keyserviceproto.GetDataKeyRequest{
			DataKeyId: keyID,
			Length:    uint32(length),
		},
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching data encryption key from Constellation KMS: %w", err)
	}

	log.Info("Data key request successful")
	return res.DataKey, nil
}

type grpcClient interface {
	GetDataKey(context.Context, *keyserviceproto.GetDataKeyRequest, *grpc.ClientConn) (*keyserviceproto.GetDataKeyResponse, error)
}

type client struct{}

func (c client) GetDataKey(ctx context.Context, req *keyserviceproto.GetDataKeyRequest, conn *grpc.ClientConn) (*keyserviceproto.GetDataKeyResponse, error) {
	return keyserviceproto.NewAPIClient(conn).GetDataKey(ctx, req)
}
