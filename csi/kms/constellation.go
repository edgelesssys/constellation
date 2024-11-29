/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kms

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/keyservice/keyserviceproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConstellationKMS is a key service to fetch volume keys.
type ConstellationKMS struct {
	endpoint string
	kms      kmsClient
}

// NewConstellationKMS initializes a ConstellationKMS.
func NewConstellationKMS(endpoint string) *ConstellationKMS {
	return &ConstellationKMS{
		endpoint: endpoint, // default: "kms.kube-system:port"
		kms:      &constellationKMSClient{},
	}
}

// GetDEK request a data encryption key derived from the Constellation's master secret.
func (k *ConstellationKMS) GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error) {
	conn, err := grpc.NewClient(k.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := k.kms.GetDataKey(
		ctx,
		&keyserviceproto.GetDataKeyRequest{
			DataKeyId: dekID,
			Length:    uint32(dekSize),
		},
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching data encryption key from Constellation KMS: %w", err)
	}

	return res.DataKey, nil
}

type kmsClient interface {
	GetDataKey(context.Context, *keyserviceproto.GetDataKeyRequest, *grpc.ClientConn) (*keyserviceproto.GetDataKeyResponse, error)
}

type constellationKMSClient struct{}

func (c *constellationKMSClient) GetDataKey(ctx context.Context, req *keyserviceproto.GetDataKeyRequest, conn *grpc.ClientConn) (*keyserviceproto.GetDataKeyResponse, error) {
	return keyserviceproto.NewAPIClient(conn).GetDataKey(ctx, req)
}
