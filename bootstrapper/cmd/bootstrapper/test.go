/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// ClusterFake behaves like a real cluster, but does not actually initialize or join Kubernetes.
type clusterFake struct{}

// InitCluster fakes bootstrapping a new cluster with the current node being the master, returning the arguments required to join the cluster.
func (c *clusterFake) InitCluster(
	context.Context, string, string, []byte, []uint32, bool, []byte, bool,
	[]byte, bool, components.Components, *logger.Logger,
) ([]byte, error) {
	return []byte{}, nil
}

// JoinCluster will fake joining the current node to an existing cluster.
func (c *clusterFake) JoinCluster(context.Context, *kubeadm.BootstrapTokenDiscovery, role.Role, string, components.Components, *logger.Logger) error {
	return nil
}

// StartKubelet starts the kubelet service.
func (c *clusterFake) StartKubelet(*logger.Logger) error {
	return nil
}

type providerMetadataFake struct{}

func (f *providerMetadataFake) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	self, err := f.Self(ctx)
	return []metadata.InstanceMetadata{self}, err
}

func (f *providerMetadataFake) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	return metadata.InstanceMetadata{
		Name:       "instanceName",
		ProviderID: "fake://instance-id",
		Role:       role.Unknown,
		VPCIP:      "192.0.2.1",
	}, nil
}

func (f *providerMetadataFake) GetLoadBalancerEndpoint(ctx context.Context) (string, error) {
	return "", nil
}

func (f *providerMetadataFake) InitSecretHash(ctx context.Context) ([]byte, error) {
	return nil, nil
}
