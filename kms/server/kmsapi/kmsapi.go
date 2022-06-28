// Package kmsapi implements an API to manage encryption keys.
package kmsapi

import (
	"context"

	"github.com/edgelesssys/constellation/internal/grpc/grpclog"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/kms"
	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// API resembles an encryption key management api server through logger, CloudKMS and proto-unimplemented server.
type API struct {
	log    *logger.Logger
	conKMS kms.CloudKMS
	kmsproto.UnimplementedAPIServer
}

// New creates a new API.
func New(log *logger.Logger, conKMS kms.CloudKMS) *API {
	return &API{
		log:    log,
		conKMS: conKMS,
	}
}

// GetDataKey returns a data key.
func (a *API) GetDataKey(ctx context.Context, in *kmsproto.GetDataKeyRequest) (*kmsproto.GetDataKeyResponse, error) {
	log := a.log.With("peerAddress", grpclog.PeerAddrFromContext(ctx))

	// Error on 0 key length
	if in.Length == 0 {
		log.Errorf("Requested key length is zero")
		return nil, status.Error(codes.InvalidArgument, "can't derive key with length zero")
	}

	// Error on empty DataKeyId
	if in.DataKeyId == "" {
		log.Errorf("No data key ID specified")
		return nil, status.Error(codes.InvalidArgument, "no data key ID specified")
	}

	key, err := a.conKMS.GetDEK(ctx, "Constellation", "key-"+in.DataKeyId, int(in.Length))
	if err != nil {
		log.With(zap.Error(err)).Errorf("Failed to get data key")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &kmsproto.GetDataKeyResponse{DataKey: key}, nil
}
