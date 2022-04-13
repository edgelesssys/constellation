package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/proto"
)

type protoClient interface {
	Connect(ip, port string, gcpPCRs, azurePCRs map[uint32][]byte) error
	Close() error
	Activate(ctx context.Context, userPublicKey, masterSecret []byte, endpoints, autoscalingNodeGroups []string, cloudServiceAccountURI string) (proto.ActivationResponseClient, error)
}
