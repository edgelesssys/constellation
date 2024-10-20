/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeVersionSpec defines the desired state of NodeVersion.
type NodeVersionSpec struct {
	// ImageReference is the image to use for all nodes.
	ImageReference string `json:"image,omitempty"`
	// ImageVersion is the CSP independent version of the image to use for all nodes.
	ImageVersion string `json:"imageVersion,omitempty"`
	// KubernetesComponentsReference is a reference to the ConfigMap containing the Kubernetes components to use for all nodes.
	KubernetesComponentsReference string `json:"kubernetesComponentsReference,omitempty"`
	// KubernetesClusterVersion is the advertised Kubernetes version of the cluster.
	KubernetesClusterVersion string `json:"kubernetesClusterVersion,omitempty"`
	// MaxNodeBudget is the maximum number of nodes that can be created simultaneously.
	MaxNodeBudget uint32 `json:"maxNodeBudget,omitempty"`
}

// NodeVersionStatus defines the observed state of NodeVersion.
type NodeVersionStatus struct {
	// Outdated is a list of nodes that are using an outdated image.
	Outdated []corev1.ObjectReference `json:"outdated,omitempty"`
	// UpToDate is a list of nodes that are using the latest image and labels.
	UpToDate []corev1.ObjectReference `json:"upToDate,omitempty"`
	// Donors is a list of outdated nodes that donate labels to heirs.
	Donors []corev1.ObjectReference `json:"donors,omitempty"`
	// Heirs is a list of nodes using the latest image that still need to inherit labels from donors.
	Heirs []corev1.ObjectReference `json:"heirs,omitempty"`
	// Mints is a list of up to date nodes that will become heirs.
	Mints []corev1.ObjectReference `json:"mints,omitempty"`
	// AwaitingAnnotation is a list of nodes that are waiting for the operator to annotate them.
	AwaitingAnnotation []corev1.ObjectReference `json:"awaitingAnnotation,omitempty"`
	// Pending is a list of pending nodes (joining or leaving the cluster).
	Pending []corev1.ObjectReference `json:"pending,omitempty"`
	// Obsolete is a list of obsolete nodes (nodes that have been created by the operator but are no longer needed).
	Obsolete []corev1.ObjectReference `json:"obsolete,omitempty"`
	// Invalid is a list of invalid nodes (nodes that cannot be processed by the operator due to missing information or transient faults).
	Invalid []corev1.ObjectReference `json:"invalid,omitempty"`
	// Budget is the amount of extra nodes that can be created as replacements for outdated nodes.
	Budget uint32 `json:"budget"`
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`
	// ActiveClusterVersionUpgrade indicates whether the cluster is currently upgrading.
	ActiveClusterVersionUpgrade bool `json:"activeclusterversionupgrade"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NodeVersion is the Schema for the nodeversions API.
type NodeVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeVersionSpec   `json:"spec,omitempty"`
	Status NodeVersionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeVersionList contains a list of NodeVersion.
type NodeVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeVersion{}, &NodeVersionList{})
}
