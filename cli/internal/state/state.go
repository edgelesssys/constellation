/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// This binary can be build from siderolabs/talos projects. Located at:
// https://github.com/siderolabs/talos/tree/master/hack/docgen
//
//go:generate docgen ./state.go ./state_doc.go Configuration

// package state defines the structure of the Constellation state file.
package state

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

const (
	// Version1 is the first version of the state file.
	Version1 = "v1"
)

// ReadFromFile reads the state file at the given path and returns the state.
func ReadFromFile(fileHandler file.Handler, path string) (*State, error) {
	state := &State{}
	if err := fileHandler.ReadYAML(path, &state); err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}
	return state, nil
}

// State describe the entire state to describe a Constellation cluster.
type State struct {
	// description: |
	//   Schema version of this state file.
	Version string `yaml:"version"`

	// TODO(msanft): Add link to self-managed infrastructure docs once existing.

	// description: |
	//   State of the cluster's cloud resources. These values are retrieved during
	//   cluster creation. In the case of self-managed infrastructure, the marked
	//   fields in this struct should be filled by the user as per
	//   https://docs.edgeless.systems/constellation/workflows/create.
	Infrastructure Infrastructure `yaml:"infrastructure"`
	// description: |
	//   DO NOT EDIT. State of the Constellation Kubernetes cluster.
	//   These values are set during cluster initialization and should not be changed.
	ClusterValues ClusterValues `yaml:"clusterValues"`
}

// ClusterValues describe the (Kubernetes) cluster state, set during initialization of the cluster.
type ClusterValues struct {
	// description: |
	//   Unique identifier of the cluster.
	ClusterID string `yaml:"clusterID"`
	// description: |
	//   Unique identifier of the owner of the cluster.
	OwnerID string `yaml:"ownerID"`
	// description: |
	//   Salt used to generate the ClusterID on the bootstrapping node.
	MeasurementSalt []byte `yaml:"measurementSalt"`
}

// Infrastructure describe the state related to the cloud resources of the cluster.
type Infrastructure struct {
	// description: |
	//   Unique identifier the cluster's cloud resources are tagged with.
	UID string `yaml:"uid"`
	// description: |
	//   Endpoint the cluster can be reached at. This value needs to be set
	//   by the user in the case of self-managed infrastructure.
	ClusterEndpoint string `yaml:"clusterEndpoint"`
	// description: |
	//   Secret used to authenticate the bootstrapping node. This value needs to be set
	//   by the user in the case of self-managed infrastructure.
	InitSecret []byte `yaml:"initSecret"`
	// description: |
	//   List of Subject Alternative Names (SANs) to add to the Kubernetes API server certificate.
	APIServerCertSANs []string `yaml:"apiServerCertSANs"`
	// description: |
	//   Name used in the cluster's named resources.
	Name string `yaml:"name"`
	// description: |
	//   Values specific to a Constellation cluster running on Azure.
	Azure *Azure `yaml:"azure,omitempty"`
	// description: |
	//   Values specific to a Constellation cluster running on GCP.
	GCP *GCP `yaml:"gcp,omitempty"`
}

// GCP describes the infra state related to GCP.
type GCP struct {
	// description: |
	//   Project ID of the GCP project the cluster is running in.
	ProjectID string `yaml:"projectID"`
	// description: |
	//   CIDR range of the cluster's nodes.
	IPCidrNode string `yaml:"ipCidrNode"`
	// description: |
	//   CIDR range of the cluster's pods.
	IPCidrPod string `yaml:"ipCidrPod"`
}

// Azure describes the infra state related to Azure.
type Azure struct {
	// description: |
	//   Resource Group the cluster's resources are placed in.
	ResourceGroup string `yaml:"resourceGroup"`
	// description: |
	//   ID of the Azure subscription the cluster is running in.
	SubscriptionID string `yaml:"subscriptionID"`
	// description: |
	//   Security group name of the cluster's resource group.
	NetworkSecurityGroupName string `yaml:"networkSecurityGroupName"`
	// description: |
	//   Name of the cluster's load balancer.
	LoadBalancerName string `yaml:"loadBalancerName"`
	// description: |
	//   ID of the UAMI the cluster's nodes are running with.
	UserAssignedIdentity string `yaml:"userAssignedIdentity"`
	// description: |
	//   MAA endpoint that can be used as a fallback for veryifying the ID key digests
	//   in the cluster's attestation report if the enforcement policy is set accordingly.
	//   Can be left empty otherwise.
	AttestationURL string `yaml:"attestationURL"`
}

// New creates a new cluster state (file).
func New() *State {
	return &State{
		Version: Version1,
	}
}

// SetInfrastructure sets the infrastructure state.
func (s *State) SetInfrastructure(infrastructure Infrastructure) *State {
	s.Infrastructure = infrastructure
	return s
}

// SetClusterValues sets the cluster values.
func (s *State) SetClusterValues(clusterValues ClusterValues) *State {
	s.ClusterValues = clusterValues
	return s
}

// WriteToFile writes the state to the given path, overwriting any existing file.
func (s *State) WriteToFile(fileHandler file.Handler, path string) error {
	if err := fileHandler.WriteYAML(path, s, file.OptMkdirAll, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// Merge merges the state information from other into the current state.
// If a field is set in both states, the value of the other state is used.
func (s *State) Merge(other *State) (*State, error) {
	if err := mergo.Merge(s, other, mergo.WithOverride); err != nil {
		return nil, fmt.Errorf("merging state file: %w", err)
	}
	return s, nil
}
