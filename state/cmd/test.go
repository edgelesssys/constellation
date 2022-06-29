package main

import (
	"context"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
)

type fakeMetadataAPI struct{}

func (f *fakeMetadataAPI) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	return []metadata.InstanceMetadata{
		{
			Name:       "instanceName",
			ProviderID: "fake://instance-id",
			Role:       role.Unknown,
			PrivateIPs: []string{"192.0.2.1"},
		},
	}, nil
}
