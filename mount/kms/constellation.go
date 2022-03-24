package kms

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConstellationKMS is a key service using the Constellation Coordinator to fetch volume keys.
type ConstellationKMS struct {
	endpoint string
	vpn      vpnClient
}

// NewConstellationKMS initializes a ConstellationKMS.
func NewConstellationKMS(coordinatorEndpoint string) *ConstellationKMS {
	return &ConstellationKMS{
		endpoint: coordinatorEndpoint, // default: "10.118.0.1:9027"
		vpn:      &constellationVPNClient{},
	}
}

// GetDEK connects to the Constellation Coordinators VPN API to request a data encryption key derived from the Constellation's master secret.
func (k *ConstellationKMS) GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error) {
	conn, err := grpc.DialContext(ctx, k.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := k.vpn.GetDataKey(
		ctx,
		&vpnproto.GetDataKeyRequest{
			DataKeyId: dekID,
			Length:    uint32(dekSize),
		},
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get data encryption key from Constellation Coordinator: %w", err)
	}

	return res.DataKey, nil
}

type vpnClient interface {
	GetDataKey(context.Context, *vpnproto.GetDataKeyRequest, *grpc.ClientConn) (*vpnproto.GetDataKeyResponse, error)
}

type constellationVPNClient struct{}

func (c *constellationVPNClient) GetDataKey(ctx context.Context, req *vpnproto.GetDataKeyRequest, conn *grpc.ClientConn) (*vpnproto.GetDataKeyResponse, error) {
	return vpnproto.NewAPIClient(conn).GetDataKey(ctx, req)
}
