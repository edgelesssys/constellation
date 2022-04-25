package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/cli/proto"
	"github.com/edgelesssys/constellation/coordinator/atls"
)

type protoClient interface {
	Connect(ip, port string, validators []atls.Validator) error
	Close() error
	Activate(ctx context.Context, userPublicKey, masterSecret []byte, nodeIPs, coordinatorIPs, autoscalingNodeGroups []string, cloudServiceAccountURI string) (proto.ActivationResponseClient, error)
}
