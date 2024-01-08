/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controlplane

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodeutil "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/node"
)

// ListControlPlaneIPs retrieves a list of VPC IPs for the control plane nodes from kubernetes.
func ListControlPlaneIPs(k8sClient client.Client) ([]string, error) {
	var nodes corev1.NodeList
	if err := k8sClient.List(context.TODO(), &nodes); err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	var ips []string
	for _, node := range nodes.Items {
		if !nodeutil.IsControlPlaneNode(&node) {
			continue
		}
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				ips = append(ips, addr.Address)
			}
		}
	}

	return ips, nil
}
