/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeImageSpec defines the desired state of NodeImage.
type NodeImageSpec struct {
	// ImageReference is the image to use for all nodes.
	ImageReference string `json:"image,omitempty"`
}

// NodeImageStatus defines the observed state of NodeImage.
type NodeImageStatus struct {
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
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NodeImage is the Schema for the nodeimages API.
type NodeImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeImageSpec   `json:"spec,omitempty"`
	Status NodeImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeImageList contains a list of NodeImage.
type NodeImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeImage{}, &NodeImageList{})
}
