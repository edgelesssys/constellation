/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConditionOutdated is used to signal outdated scaling groups.
	ConditionOutdated = "Outdated"
)

// ScalingGroupSpec defines the desired state of ScalingGroup
type ScalingGroupSpec struct {
	// NodeImage is the name of the NodeImage resource.
	NodeImage string `json:"nodeImage,omitempty"`
	// GroupID is the CSP specific, canonical identifier of a scaling group.
	GroupID string `json:"groupId,omitempty"`
	// Autoscaling specifies wether the scaling group should automatically scale using the cluster-autoscaler.
	Autoscaling bool `json:"autoscaling,omitempty"`
}

// ScalingGroupStatus defines the observed state of ScalingGroup
type ScalingGroupStatus struct {
	// ImageReference is the image currently used for newly created nodes in this scaling group.
	ImageReference string `json:"imageReference,omitempty"`
	// Conditions represent the latest available observations of an object's state.
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ScalingGroup is the Schema for the scalinggroups API
type ScalingGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScalingGroupSpec   `json:"spec,omitempty"`
	Status ScalingGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScalingGroupList contains a list of ScalingGroup
type ScalingGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalingGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalingGroup{}, &ScalingGroupList{})
}
