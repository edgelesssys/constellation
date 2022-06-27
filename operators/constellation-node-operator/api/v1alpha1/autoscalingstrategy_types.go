
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutoscalingStrategySpec defines the desired state of AutoscalingStrategy
type AutoscalingStrategySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AutoscalingStrategy. Edit autoscalingstrategy_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AutoscalingStrategyStatus defines the observed state of AutoscalingStrategy
type AutoscalingStrategyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
