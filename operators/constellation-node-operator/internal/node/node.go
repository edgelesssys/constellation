/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package node

import (
	"errors"
	"regexp"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const controlPlaneRoleLabel = "node-role.kubernetes.io/control-plane"

var reservedHostRegex = regexp.MustCompile(`^(.+\.|)(kubernetes|k8s)\.io(/.*)?$`)

// Ready checks if a kubernetes node has the `NodeReady` condition set to true.
func Ready(node *corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// VPCIP returns the VPC IP of a node.
func VPCIP(node *corev1.Node) (string, error) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", errors.New("no VPC IP found")
}

// IsControlPlaneNode returns true if the node is a control plane node.
func IsControlPlaneNode(node *corev1.Node) bool {
	_, ok := node.Labels[controlPlaneRoleLabel]
	return ok
}

// FindPending searches for a pending node that matches a node.
// The pending node has to have the goal to join the cluster and be reported as ready be the CSP.
// if the node is not found, nil is returned.
func FindPending(pendingNodes []updatev1alpha1.PendingNode, node *corev1.Node) *updatev1alpha1.PendingNode {
	if node == nil {
		return nil
	}
	for _, pendingNode := range pendingNodes {
		if pendingNode.Spec.Goal == updatev1alpha1.NodeGoalJoin && pendingNode.Spec.NodeName == node.Name && pendingNode.Status.CSPNodeState == updatev1alpha1.NodeStateReady {
			return &pendingNode
		}
	}
	return nil
}

// FilterLabels removes reserved node labels from a map of labels.
// reference: https://kubernetes.io/docs/reference/labels-annotations-taints/ .
func FilterLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for key, val := range labels {
		if reservedHostRegex.MatchString(key) {
			continue
		}
		result[key] = val
	}
	return result
}
