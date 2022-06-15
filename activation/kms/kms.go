package kms

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
)

// Client interacts with Constellation's key management service.
type Client struct {
	endpoint string
	grpc     grpcClient
}

// New creates a new KMS.
func New(endpoint string) Client {
	return Client{
		endpoint: endpoint,
		grpc:     client{},
	}
}

// GetDEK returns a data encryption key for the given UUID.
func (c Client) GetDataKey(ctx context.Context, uuid string, length int) ([]byte, error) {
	// TODO: update credentials if we enable aTLS on the KMS
	// For now this is fine since traffic is only routed through the Constellation cluster
	conn, err := grpc.DialContext(ctx, c.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	klog.V(6).Infof("GetDataKey: connecting to KMS at %s", c.endpoint)
	res, err := c.grpc.GetDataKey(
		ctx,
		&kmsproto.GetDataKeyRequest{
			DataKeyId: uuid,
			Length:    uint32(length),
		},
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching data encryption key from Constellation KMS: %w", err)
	}

	return res.DataKey, nil
}

type grpcClient interface {
	GetDataKey(context.Context, *kmsproto.GetDataKeyRequest, *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error)
}

type client struct{}

func (c client) GetDataKey(ctx context.Context, req *kmsproto.GetDataKeyRequest, conn *grpc.ClientConn) (*kmsproto.GetDataKeyResponse, error) {
	return kmsproto.NewAPIClient(conn).GetDataKey(ctx, req)
}
