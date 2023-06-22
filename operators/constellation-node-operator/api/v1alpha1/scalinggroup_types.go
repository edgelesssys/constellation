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

	// WorkerRole is used to signal worker scaling groups.
	WorkerRole NodeRole = "Worker"
	// ControlPlaneRole is used to signal control plane scaling groups.
	ControlPlaneRole NodeRole = "ControlPlane"
)

// ScalingGroupSpec defines the desired state of ScalingGroup.
type ScalingGroupSpec struct {
	// NodeVersion is the name of the NodeVersion resource.
	NodeVersion string `json:"nodeImage,omitempty"`
	// GroupID is the CSP specific, canonical identifier of a scaling group.
	GroupID string `json:"groupId,omitempty"`
	// AutoscalerGroupName is name that is expected by the autoscaler.
	AutoscalerGroupName string `json:"autoscalerGroupName,omitempty"`
	// NodeGroupName is the human friendly name of the node group as defined in the Constellation configuration.
	NodeGroupName string `json:"nodeGroupName,omitempty"`
	// Autoscaling specifies wether the scaling group should automatically scale using the cluster-autoscaler.
	Autoscaling bool `json:"autoscaling,omitempty"`
	// Min is the minimum number of nodes in the scaling group (used by cluster-autoscaler).
	Min int32 `json:"min,omitempty"`
	// Max is the maximum number of autoscaled nodes in the scaling group (used by cluster-autoscaler).
	Max int32 `json:"max,omitempty"`
	// Role is the role of the nodes in the scaling group.
	Role NodeRole `json:"role,omitempty"`
}

// NodeRole is the role of a node.
// +kubebuilder:validation:Enum=Worker;ControlPlane
type NodeRole string

// ScalingGroupStatus defines the observed state of ScalingGroup.
type ScalingGroupStatus struct {
	// ImageReference is the image currently used for newly created nodes in this scaling group.
	ImageReference string `json:"imageReference,omitempty"`
	// Conditions represent the latest available observations of an object's state.
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ScalingGroup is the Schema for the scalinggroups API.
type ScalingGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScalingGroupSpec   `json:"spec,omitempty"`
	Status ScalingGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScalingGroupList contains a list of ScalingGroup.
type ScalingGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalingGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalingGroup{}, &ScalingGroupList{})
}
