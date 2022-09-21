/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metadata

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// InstanceMetadata describes metadata of a peer.
type InstanceMetadata struct {
	Name       string
	ProviderID string
	Role       role.Role
	// VPCIP is the primary IP address of the instance in the VPC.
	VPCIP string
	// PublicIP is the primary public IP of the instance, if available, empty string otherwise.
	PublicIP      string
	AliasIPRanges []string
	// SSHKeys maps usernames to ssh public keys.
	SSHKeys map[string][]string
}

type InstanceSelfer interface {
	// Self retrieves the current instance.
	Self(ctx context.Context) (InstanceMetadata, error)
}

type InstanceLister interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]InstanceMetadata, error)
}

// JoinServiceEndpoints returns the list of endpoints for the join service, which are running on the control plane nodes.
func JoinServiceEndpoints(ctx context.Context, lister InstanceLister) ([]string, error) {
	instances, err := lister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instances list from cloud provider: %w", err)
	}
	joinEndpoints := []string{}
	for _, instance := range instances {
		if instance.Role == role.ControlPlane {
			if instance.VPCIP != "" {
				joinEndpoints = append(joinEndpoints, net.JoinHostPort(instance.VPCIP, strconv.Itoa(constants.JoinServiceNodePort)))
			}
		}
	}

	return joinEndpoints, nil
}
