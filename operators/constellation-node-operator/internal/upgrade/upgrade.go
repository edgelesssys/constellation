/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"fmt"
	"net"

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"

	"github.com/edgelesssys/constellation/v2/upgrade-agent/upgradeproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a client for the upgrade agent.
type Client struct {
	dialer Dialer
}

// NewClient creates a new upgrade agent client.
func NewClient() *Client {
	return &Client{
		dialer: &net.Dialer{},
	}
}

// Upgrade upgrades the Constellation node to the given Kubernetes version.
func (c *Client) Upgrade(ctx context.Context, kubernetesComponents components.Components, WantedKubernetesVersion string) error {
	conn, err := grpc.DialContext(ctx, mainconstants.UpgradeAgentMountPath, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(
			func(ctx context.Context, addr string) (net.Conn, error) {
				return c.dialer.DialContext(ctx, "unix", addr)
			},
		))
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	// While we're transitioning between version 2.13 and 2.14, we need to
	// expect an upgrade-agent that does not yet understand the
	// KubernetesComponents proto field. Therefore, we pass the kubeadm
	// component twice: once via KubeadmUrl/KubeadmHash, once as part of the
	// kubernetesComponents argument.
	kubeadm, err := kubernetesComponents.GetKubeadmComponent()
	if err != nil {
		return fmt.Errorf("expected a kubeadm Component: %w", err)
	}
	protoClient := upgradeproto.NewUpdateClient(conn)
	_, err = protoClient.ExecuteUpdate(ctx, &upgradeproto.ExecuteUpdateRequest{
		// TODO(burgerdev): remove these fields after releasing 2.14.
		// %< ---------------------------------
		KubeadmUrl:  kubeadm.Url,
		KubeadmHash: kubeadm.Hash,
		// %< ---------------------------------
		WantedKubernetesVersion: WantedKubernetesVersion,
		KubernetesComponents:    kubernetesComponents,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}
	return nil
}

// Dialer is a dialer for the upgrade agent.
type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}
