package cmd

import (
	"context"
	"io"

	"github.com/edgelesssys/constellation/coordinator/atls"
)

type recoveryClient interface {
	Connect(ip, port string, validators []atls.Validator) error
	PushStateDiskKey(ctx context.Context, stateDiskKey []byte) error
	io.Closer
}
