
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PendingNodeSpec defines the desired state of PendingNode
type PendingNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PendingNode. Edit pendingnode_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PendingNodeStatus defines the observed state of PendingNode
type PendingNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// PendingNode is the Schema for the pendingnodes API
type PendingNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PendingNodeSpec   `json:"spec,omitempty"`
	Status PendingNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PendingNodeList contains a list of PendingNode
type PendingNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PendingNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PendingNode{}, &PendingNodeList{})
}
