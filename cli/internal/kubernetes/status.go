/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"fmt"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// TargetVersions bundles version information about the target versions of a cluster.
type TargetVersions struct {
	// image version
	image string
	// CSP specific path to the image
	imageReference string
	// kubernetes version
	kubernetes string
}

// NewTargetVersions returns the target versions for the cluster.
func NewTargetVersions(nodeVersion updatev1alpha1.NodeVersion) (TargetVersions, error) {
	return TargetVersions{
		image:          nodeVersion.Spec.ImageVersion,
		imageReference: nodeVersion.Spec.ImageReference,
		kubernetes:     nodeVersion.Spec.KubernetesClusterVersion,
	}, nil
}

// Image return the image version.
func (c *TargetVersions) Image() string {
	return c.image
}

// ImagePath return the image path.
func (c *TargetVersions) ImagePath() string {
	return c.imageReference
}

// Kubernetes return the Kubernetes version.
func (c *TargetVersions) Kubernetes() string {
	return c.kubernetes
}

// ClusterStatus returns a map from node name to NodeStatus.
func ClusterStatus(ctx context.Context, kubeclient kubeClient) (map[string]NodeStatus, error) {
	nodes, err := kubeclient.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting nodes: %w", err)
	}

	clusterStatus := map[string]NodeStatus{}
	for _, node := range nodes {
		clusterStatus[node.ObjectMeta.Name] = NewNodeStatus(node)
	}

	return clusterStatus, nil
}

// NodeStatus bundles status information about a node.
type NodeStatus struct {
	kubeletVersion string
	imageVersion   string
}

// NewNodeStatus returns a new NodeStatus.
func NewNodeStatus(node corev1.Node) NodeStatus {
	return NodeStatus{
		kubeletVersion: node.Status.NodeInfo.KubeletVersion,
		imageVersion:   node.ObjectMeta.Annotations["constellation.edgeless.systems/node-image"],
	}
}

// KubeletVersion returns the kubelet version of the node.
func (n *NodeStatus) KubeletVersion() string {
	return n.kubeletVersion
}

// ImageVersion returns the node image of the node.
func (n *NodeStatus) ImageVersion() string {
	return n.imageVersion
}

type kubeClient interface {
	GetNodes(ctx context.Context) ([]corev1.Node, error)
}
