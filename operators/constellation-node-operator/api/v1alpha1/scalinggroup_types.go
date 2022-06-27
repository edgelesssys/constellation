
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScalingGroupSpec defines the desired state of ScalingGroup
type ScalingGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ScalingGroup. Edit scalinggroup_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ScalingGroupStatus defines the observed state of ScalingGroup
type ScalingGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
