/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JoiningNodeSpec defines the components hash which the node should be annotated with.
type JoiningNodeSpec struct {
	// Name of the node expected to join.
	Name string `json:"name,omitempty"`
	// ComponentsHash is the hash of the components that were sent to the node by the join service.
	ComponentsHash string `json:"componentshash,omitempty"`
	// IsControlPlane is true if the node is a control plane node.
	IsControlPlane bool `json:"iscontrolplane,omitempty"`
	// Deadline is the time after which the joining node is considered to have failed.
	Deadline *metav1.Time `json:"deadline,omitempty"`
}

// JoiningNodeStatus defines the observed state of JoiningNode.
type JoiningNodeStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// JoiningNode is the Schema for the joiningnodes API.
type JoiningNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JoiningNodeSpec   `json:"spec,omitempty"`
	Status JoiningNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JoiningNodeList contains a list of JoiningNodes.
type JoiningNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JoiningNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JoiningNode{}, &JoiningNodeList{})
}
