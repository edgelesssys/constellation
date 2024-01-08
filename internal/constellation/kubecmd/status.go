/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubecmd

import (
	"fmt"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// NodeVersion bundles version information of a Constellation cluster.
type NodeVersion struct {
	imageVersion      string
	imageReference    string
	kubernetesVersion string
	clusterStatus     string
}

// NewNodeVersion returns the target versions for the cluster.
func NewNodeVersion(nodeVersion updatev1alpha1.NodeVersion) (NodeVersion, error) {
	if len(nodeVersion.Status.Conditions) != 1 {
		return NodeVersion{}, fmt.Errorf("expected exactly one condition, got %d", len(nodeVersion.Status.Conditions))
	}
	return NodeVersion{
		imageVersion:      nodeVersion.Spec.ImageVersion,
		imageReference:    nodeVersion.Spec.ImageReference,
		kubernetesVersion: nodeVersion.Spec.KubernetesClusterVersion,
		clusterStatus:     nodeVersion.Status.Conditions[0].Message,
	}, nil
}

// ImageVersion is the version of the image running on a node.
func (n NodeVersion) ImageVersion() string {
	return n.imageVersion
}

// ImageReference is a CSP specific path to the image.
func (n NodeVersion) ImageReference() string {
	return n.imageReference
}

// KubernetesVersion is the Kubernetes version running on a node.
func (n NodeVersion) KubernetesVersion() string {
	return n.kubernetesVersion
}

// ClusterStatus is a string describing the status of the cluster.
func (n NodeVersion) ClusterStatus() string {
	return n.clusterStatus
}

// NodeStatus bundles status information about a Kubernetes node.
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

func updateNodeVersions(newNodeVersion updatev1alpha1.NodeVersion, node *updatev1alpha1.NodeVersion) {
	if newNodeVersion.Spec.ImageVersion != "" {
		node.Spec.ImageVersion = newNodeVersion.Spec.ImageVersion
	}
	if newNodeVersion.Spec.ImageReference != "" {
		node.Spec.ImageReference = newNodeVersion.Spec.ImageReference
	}
	if newNodeVersion.Spec.KubernetesComponentsReference != "" {
		node.Spec.KubernetesComponentsReference = newNodeVersion.Spec.KubernetesComponentsReference
	}
	if newNodeVersion.Spec.KubernetesClusterVersion != "" {
		node.Spec.KubernetesClusterVersion = newNodeVersion.Spec.KubernetesClusterVersion
	}
}
