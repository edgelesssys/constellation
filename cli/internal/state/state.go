/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package state defines the structure of the Constellation state file.
package state

const (
	// Version1 is the first version of the state file.
	Version1 = "v1"
)

// State describe the entire state to describe a Constellation cluster.
type State struct {
	Version        string         `yaml:"version"`
	Infrastructure Infrastructure `yaml:"infrastructure"`
}

// NewState creates a new state with the given infrastructure.
func NewState(Infrastructure Infrastructure) State {
	return State{
		Version:        Version1,
		Infrastructure: Infrastructure,
	}
}

// Infrastructure describe the state related to the cloud resources of the cluster.
type Infrastructure struct {
	UID               string   `yaml:"uid"`
	ClusterEndpoint   string   `yaml:"clusterEndpoint"`
	InitSecret        string   `yaml:"initSecret"`
	APIServerCertSANs []string `yaml:"apiServerCertSANs"`
	Azure             *Azure   `yaml:"azure,omitempty"`
	GCP               *GCP     `yaml:"gcp,omitempty"`
}

// GCP describes the infra state related to GCP.
type GCP struct {
	ProjectID  string `yaml:"projectID"`
	IPCidrNode string `yaml:"ipCidrNode"`
	IPCidrPod  string `yaml:"ipCidrPod"`
}

// Azure describes the infra state related to Azure.
type Azure struct {
	ResourceGroup            string `yaml:"resourceGroup"`
	SubscriptionID           string `yaml:"subscriptionID"`
	NetworkSecurityGroupName string `yaml:"networkSecurityGroupName"`
	LoadBalancerName         string `yaml:"loadBalancerName"`
	UserAssignedIdentity     string `yaml:"userAssignedIdentity"`
	AttestationURL           string `yaml:"attestationURL"`
}
