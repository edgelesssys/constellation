package kms

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/kms/kmsproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConstellationKMS is a key service using the Constellation Coordinator to fetch volume keys.
type ConstellationKMS struct {
	endpoint string
	kms      kmsClient
}

// NewConstellationKMS initializes a ConstellationKMS.
func NewConstellationKMS(coordinatorEndpoint string) *ConstellationKMS {
	return &ConstellationKMS{
		endpoint: coordinatorEndpoint, // default: "kms.kube-system:9000"
		kms:      &constellationKMSClient{},
	}
}

// GetDEK connects to the Constellation Coordinators VPN API to request a data encryption key derived from the Constellation's master secret.
func (k *ConstellationKMS) GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error) {
	conn, err := grpc.DialContext(ctx, k.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := k.kms.GetDataKey(
		ctx,
		&kmsproto.GetDataKeyRequest{
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
	GetDataKey(context.Context, *kmsproto.GetDataKeyRequest, *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error)
}

type constellationKMSClient struct{}

func (c *constellationKMSClient) GetDataKey(ctx context.Context, req *kmsproto.GetDataKeyRequest, conn *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error) {
	return kmsproto.NewAPIClient(conn).GetDataKey(ctx, req)
}
