/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoscalingStrategySpec defines the desired state of AutoscalingStrategy
type AutoscalingStrategySpec struct {
	// Enabled defines whether cluster autoscaling should be enabled or not.
	Enabled bool `json:"enabled"`
	// DeploymentName defines the name of the autoscaler deployment.
	DeploymentName string `json:"deploymentName"`
	// DeploymentNamespace defines the namespace of the autoscaler deployment.
	DeploymentNamespace string `json:"deploymentNamespace"`
}

// AutoscalingStrategyStatus defines the observed state of AutoscalingStrategy
type AutoscalingStrategyStatus struct {
	// Enabled shows whether cluster autoscaling is currently enabled or not.
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Replicas is the number of replicas for the autoscaler deployment.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// AutoscalingStrategy is the Schema for the autoscalingstrategies API
type AutoscalingStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoscalingStrategySpec   `json:"spec,omitempty"`
	Status AutoscalingStrategyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutoscalingStrategyList contains a list of AutoscalingStrategy
type AutoscalingStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoscalingStrategy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoscalingStrategy{}, &AutoscalingStrategyList{})
}
