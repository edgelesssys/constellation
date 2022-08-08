package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NodeGoalJoin is the goal to join the cluster.
	NodeGoalJoin PendingNodeGoal = "Join"
	// NodeGoalLeave is the goal to leave the cluster and terminate the node.
	NodeGoalLeave PendingNodeGoal = "Leave"

	// NodeStateUnknown is the default state of the node if no information is available.
	NodeStateUnknown CSPNodeState = "Unknown"
	// NodeStateCreating is the state of the node when it is being created.
	NodeStateCreating CSPNodeState = "Creating"
	// NodeStateReady is the state of the node when it is ready to use.
	// This state is reached when the CSP reports a node to be ready.
	// This does not guarantee that a node has already joined the cluster.
	NodeStateReady CSPNodeState = "Ready"
	// NodeStateStopped is the state of the node when not running temporarily.
	NodeStateStopped CSPNodeState = "Stopped"
	// NodeStateTerminating is the state of the node when it is being terminated.
	NodeStateTerminating CSPNodeState = "Terminating"
	// NodeStateTerminated is the state of the node when it is terminated.
	NodeStateTerminated CSPNodeState = "Terminated"
	// NodeStateFailed is the state of the node when it encounters an unrecoverable error.
	NodeStateFailed CSPNodeState = "Failed"
)

// PendingNodeGoal is the desired state of PendingNode.
// Only one of the following goals may be specified.
// +kubebuilder:validation:Enum=Join;Leave
type PendingNodeGoal string

// CSPNodeState is the state of a Node in the cloud.
// Only one of the following states may be specified.
// +kubebuilder:validation:Enum=Unknown;Creating;Ready;Stopped;Terminating;Terminated;Failed
type CSPNodeState string

// PendingNodeSpec defines the desired state of PendingNode
type PendingNodeSpec struct {
	// ProviderID is the provider ID of the node.
	ProviderID string `json:"providerID,omitempty"`
	// ScalingGroupID is the ID of the group that this node shall be part of.
	ScalingGroupID string `json:"groupID,omitempty"`
	// NodeName is the kubernetes internal name of the node.
	NodeName string `json:"nodeName,omitempty"`
	// Goal is the goal of the pending state.
	Goal PendingNodeGoal `json:"goal,omitempty"`
	// Deadline is the deadline for reaching the goal state.
	// Joining nodes will be terminated if the deadline is exceeded.
	// Leaving nodes will remain as unschedulable to prevent data loss.
	// If not specified, the node may remain in the pending state indefinitely.
	// +optional
	Deadline *metav1.Time `json:"deadline,omitempty"`
}

// PendingNodeStatus defines the observed state of PendingNode
type PendingNodeStatus struct {
	// CSPNodeState is the state of the node in the cloud.
	CSPNodeState `json:"cspState,omitempty"`
	// ReachedGoal is true if the node has reached the goal state.
	ReachedGoal bool `json:"reachedGoal,omitempty"`
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
