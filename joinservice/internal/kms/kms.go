/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kms

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/kms/kmsproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client interacts with Constellation's key management service.
type Client struct {
	log      *logger.Logger
	endpoint string
	grpc     grpcClient
}

// New creates a new KMS.
func New(log *logger.Logger, endpoint string) Client {
	return Client{
		log:      log,
		endpoint: endpoint,
		grpc:     client{},
	}
}

// GetDataKey returns a data encryption key for the given UUID.
func (c Client) GetDataKey(ctx context.Context, keyID string, length int) ([]byte, error) {
	log := c.log.With(zap.String("keyID", keyID), zap.String("endpoint", c.endpoint))
	// the KMS does not use aTLS since traffic is only routed through the Constellation cluster
	// cluster internal connections are considered trustworthy
	log.Infof("Connecting to KMS at %s", c.endpoint)
	conn, err := grpc.DialContext(ctx, c.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	log.Infof("Requesting data key")
	res, err := c.grpc.GetDataKey(
		ctx,
		&kmsproto.GetDataKeyRequest{
			DataKeyId: keyID,
			Length:    uint32(length),
		},
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching data encryption key from Constellation KMS: %w", err)
	}

	log.Infof("Data key request successful")
	return res.DataKey, nil
}

type grpcClient interface {
	GetDataKey(context.Context, *kmsproto.GetDataKeyRequest, *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error)
}

type client struct{}

func (c client) GetDataKey(ctx context.Context, req *kmsproto.GetDataKeyRequest, conn *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error) {
	return kmsproto.NewAPIClient(conn).GetDataKey(ctx, req)
}
