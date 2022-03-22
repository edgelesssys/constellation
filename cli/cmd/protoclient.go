package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/proto"
)

type protoClient interface {
	Connect(ip string, port string) error
	Close() error
	Activate(ctx context.Context, userPublicKey, masterSecret []byte, endpoints, autoscalingNodeGroups []string, cloudServiceAccountURI string) (proto.ActivationResponseClient, error)
}
