/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package state defines the structure of the Constellation state file.
package state

// State describe the entire state to describe a Constellation cluster.
type State struct {
	Version        string         `yaml:"version"`
	Infrastructure Infrastructure `yaml:"infrastructure"`
}

// Infrastructure describe the state related to the cloud resources of the cluster.
type Infrastructure struct {
	UID               string   `yaml:"uid"`
	PublicIP          string   `yaml:"publicIP"`
	InitSecret        string   `yaml:"initSecret"`
	APIServerCertSANs []string `yaml:"apiServerCertSANs"`
	Azure             *Azure   `yaml:"azure"`
	GCP               *GCP     `yaml:"gcp"`
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
