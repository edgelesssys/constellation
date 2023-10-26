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
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/validation"
)

const (
	// Version1 is the first version of the state file.
	Version1 = "v1"
)

// ReadFromFile reads the state file at the given path and validates it.
// If the state file is valid, the state is returned. Otherwise, an error
// describing why the validation failed is returned.
func ReadFromFile(fileHandler file.Handler, path string) (*State, error) {
	state := &State{}
	if err := fileHandler.ReadYAML(path, &state); err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	v := validation.NewValidator()
	if err := v.Validate(state, validation.ValidateOptions{}); err != nil {
		return nil, fmt.Errorf("validating state file: %w", err)
	}

	return state, nil
}

// CreateOrRead reads the state file at the given path, if it exists, and returns the state.
// If the file does not exist, a new state is created and written to disk.
func CreateOrRead(fileHandler file.Handler, path string) (*State, error) {
	state, err := ReadFromFile(fileHandler, path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading state file: %w", err)
		}
		newState := New()
		return newState, newState.WriteToFile(fileHandler, path)
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
	MeasurementSalt HexBytes `yaml:"measurementSalt"`
}

// Infrastructure describe the state related to the cloud resources of the cluster.
type Infrastructure struct {
	// description: |
	//   Unique identifier the cluster's cloud resources are tagged with.
	UID string `yaml:"uid"`
	// description: |
	//   Endpoint the cluster can be reached at. This is the endpoint that is being used by the CLI.
	ClusterEndpoint string `yaml:"clusterEndpoint"`
	// description: |
	//   The Cluster uses to reach itself. This might differ from the ClusterEndpoint in case e.g.,
	//   an internal load balancer is used.
	InClusterEndpoint string `yaml:"inClusterEndpoint"`
	// description: |
	//   Secret used to authenticate the bootstrapping node.
	InitSecret HexBytes `yaml:"initSecret"`
	// description: |
	//   List of Subject Alternative Names (SANs) to add to the Kubernetes API server certificate.
	// 	 If no SANs should be added, this field can be left empty.
	APIServerCertSANs []string `yaml:"apiServerCertSANs"`
	// description: |
	//   Name used in the cluster's named resources.
	Name string `yaml:"name"`
	// description: |
	//   CIDR range of the cluster's nodes.
	IPCidrNode string `yaml:"ipCidrNode"`
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

// Constraints implements the "Validatable" interface.
func (s *State) Constraints() []*validation.Constraint {
	constraints := []*validation.Constraint{
		// state version needs to be accepted by the parsing CLI.
		validation.OneOf(s.Version, []string{Version1}).
			WithFieldTrace(s, &s.Version),
		// infrastructure should either be
		// - completely empty, in the case of Constellation-managed infrastructure, or
		// - valid already, in the case of self-managed infrastructure.
		validation.Or(
			// "valid already":
			// all constraints on the infrastructure struct must be satisfied.
			s.Infrastructure.Constraints(s),
			// "completely empty":
			// As the infrastructure struct contains slices, we cannot use the
			// Empty constraint on the entire struct. Instead, we need to check
			// each field individually.
			validation.And(
				validation.EvaluateAll,
				validation.Empty(s.Infrastructure.UID).
					WithFieldTrace(s, &s.Infrastructure.UID),
				validation.Empty(s.Infrastructure.ClusterEndpoint).
					WithFieldTrace(s, &s.Infrastructure.ClusterEndpoint),
				validation.Empty(s.Infrastructure.InClusterEndpoint).
					WithFieldTrace(s, &s.Infrastructure.InClusterEndpoint),
				validation.Empty(s.Infrastructure.Name).
					WithFieldTrace(s, &s.Infrastructure.Name),
				validation.Empty(s.Infrastructure.IPCidrNode).
					WithFieldTrace(s, &s.Infrastructure.IPCidrNode),
				validation.EmptySlice(s.Infrastructure.APIServerCertSANs).
					WithFieldTrace(s, &s.Infrastructure.APIServerCertSANs),
				validation.EmptySlice(s.Infrastructure.InitSecret).
					WithFieldTrace(s, &s.Infrastructure.InitSecret),
			),
		),
		// clusterValues should either be
		// - completely empty, before init, or
		// - completely valid, after init.
		validation.Or(
			// "completely valid":
			// all constraints on the cluster vlaues struct must be satisfied.
			s.ClusterValues.Constraints(s),
			// "completely empty":
			// As the clusterValues struct contains slices, we cannot use the
			// Empty constraint on the entire struct. Instead, we need to check
			// each field individually.
			validation.And(
				validation.EvaluateAll,
				validation.Empty(s.ClusterValues.ClusterID).
					WithFieldTrace(s, &s.ClusterValues.ClusterID),
				validation.Empty(s.ClusterValues.OwnerID).
					WithFieldTrace(s, &s.ClusterValues.OwnerID),
				validation.EmptySlice(s.ClusterValues.MeasurementSalt).
					WithFieldTrace(s, &s.ClusterValues.MeasurementSalt),
			),
		),
	}
	return constraints
}

// Constraints returns the constraints on the (not-empty) infrastructure state, ANDed to 1 constraint.
// The state is passed as an argument to allow for full JSONPaths for the field
// traces, starting from the top-level document (i.e. the state).
func (i *Infrastructure) Constraints(s *State) *validation.Constraint {
	return validation.And(
		validation.EvaluateAll,
		// out-of-cluster endpoint needs to be a valid DNS name or IP address.
		validation.Or(
			validation.DNSName(i.ClusterEndpoint).
				WithFieldTrace(s, &s.Infrastructure.ClusterEndpoint),
			validation.IPAddress(i.ClusterEndpoint).
				WithFieldTrace(s, &s.Infrastructure.ClusterEndpoint),
		),
		// in-cluster endpoint needs to be a valid DNS name or IP address.
		validation.Or(
			validation.DNSName(i.InClusterEndpoint).
				WithFieldTrace(s, &s.Infrastructure.InClusterEndpoint),
			validation.IPAddress(i.InClusterEndpoint).
				WithFieldTrace(s, &s.Infrastructure.InClusterEndpoint),
		),
		// all SANs need to be valid or the slice needs to be empty.
		validation.Or(
			validation.EmptySlice(i.APIServerCertSANs).
				WithFieldTrace(s, &s.Infrastructure.APIServerCertSANs),
			validation.All(i.APIServerCertSANs,
				func(i int, san string) *validation.Constraint {
					return validation.Or(
						validation.DNSName(san).
							WithFieldTrace(s, &s.Infrastructure.APIServerCertSANs[i]),
						validation.IPAddress(san).
							WithFieldTrace(s, &s.Infrastructure.APIServerCertSANs[i]),
					)
				},
			),
		),
		// Node IP Cidr needs to be a valid CIDR range.
		validation.CIDR(i.IPCidrNode).
			WithFieldTrace(s, &s.Infrastructure.IPCidrNode),
		// UID needs to be filled.
		validation.NotEmpty(i.UID).
			WithFieldTrace(s, &s.Infrastructure.UID),
		// Name needs to be filled.
		validation.NotEmpty(i.Name).
			WithFieldTrace(s, &s.Infrastructure.Name),
		// GCP values need to be nil, empty, or valid.
		validation.Or(
			validation.Or(
				// nil.
				validation.Equal(i.GCP, nil).
					WithFieldTrace(s, &s.Infrastructure.GCP),
				// empty.
				validation.IfNotNil(
					i.GCP,
					func() *validation.Constraint {
						return validation.And(
							validation.EvaluateAll,
							validation.Empty(i.GCP.ProjectID).
								WithFieldTrace(s, &s.Infrastructure.GCP.ProjectID),
							validation.Empty(i.GCP.IPCidrPod).
								WithFieldTrace(s, &s.Infrastructure.GCP.IPCidrPod),
						)
					},
				),
			),
			// valid.
			validation.IfNotNil(
				i.GCP,
				func() *validation.Constraint {
					return validation.And(
						validation.EvaluateAll,
						// ProjectID needs to be filled.
						validation.NotEmpty(i.GCP.ProjectID).
							WithFieldTrace(s, &s.Infrastructure.GCP.ProjectID),
						// Pod IP Cidr needs to be a valid CIDR range.
						validation.CIDR(i.GCP.IPCidrPod).
							WithFieldTrace(s, &s.Infrastructure.GCP.IPCidrPod),
					)
				},
			),
		),
		// Azure values need to be nil, empty, or valid.
		validation.Or(
			validation.Or(
				// nil.
				validation.Equal(i.Azure, nil).
					WithFieldTrace(s, &s.Infrastructure.Azure),
				// empty.
				validation.IfNotNil(
					i.Azure,
					func() *validation.Constraint {
						return validation.And(
							validation.EvaluateAll,
							validation.Empty(i.Azure.ResourceGroup).
								WithFieldTrace(s, &s.Infrastructure.Azure.ResourceGroup),
							validation.Empty(i.Azure.SubscriptionID).
								WithFieldTrace(s, &s.Infrastructure.Azure.SubscriptionID),
							validation.Empty(i.Azure.NetworkSecurityGroupName).
								WithFieldTrace(s, &s.Infrastructure.Azure.NetworkSecurityGroupName),
							validation.Empty(i.Azure.LoadBalancerName).
								WithFieldTrace(s, &s.Infrastructure.Azure.LoadBalancerName),
							validation.Empty(i.Azure.UserAssignedIdentity).
								WithFieldTrace(s, &s.Infrastructure.Azure.UserAssignedIdentity),
							validation.Empty(i.Azure.AttestationURL).
								WithFieldTrace(s, &s.Infrastructure.Azure.AttestationURL),
						)
					},
				),
			),
			// valid.
			validation.IfNotNil(
				i.Azure,
				func() *validation.Constraint {
					return validation.And(
						validation.EvaluateAll,
						validation.NotEmpty(i.Azure.ResourceGroup).
							WithFieldTrace(s, &s.Infrastructure.Azure.ResourceGroup),
						validation.NotEmpty(i.Azure.SubscriptionID).
							WithFieldTrace(s, &s.Infrastructure.Azure.SubscriptionID),
						validation.NotEmpty(i.Azure.NetworkSecurityGroupName).
							WithFieldTrace(s, &s.Infrastructure.Azure.NetworkSecurityGroupName),
						validation.NotEmpty(i.Azure.LoadBalancerName).
							WithFieldTrace(s, &s.Infrastructure.Azure.LoadBalancerName),
						validation.NotEmpty(i.Azure.UserAssignedIdentity).
							WithFieldTrace(s, &s.Infrastructure.Azure.UserAssignedIdentity),
						validation.NotEmpty(i.Azure.AttestationURL).
							WithFieldTrace(s, &s.Infrastructure.Azure.AttestationURL),
					)
				},
			),
		),
	)
}

// Constraints returns the constraints on the (not-empty) Azure infrastructure values, ANDed to 1 constraint.
// The state is passed as an argument to allow for full JSONPaths for the field
// traces, starting from the top-level document (i.e. the state).
func (c *Azure) Constraints(s *State) *validation.Constraint {
	return validation.And(
		validation.EvaluateAll,
	)
}

// Constraints returns the constraints on the (not-empty) cluster values, ANDed to 1 constraint.
// The state is passed as an argument to allow for full JSONPaths for the field
// traces, starting from the top-level document (i.e. the state).
func (c *ClusterValues) Constraints(s *State) *validation.Constraint {
	return validation.And(
		validation.EvaluateAll,
		// ClusterID needs to be filled.
		validation.NotEmpty(c.ClusterID).
			WithFieldTrace(s, &s.ClusterValues.ClusterID),
		// OwnerID needs to be filled.
		validation.NotEmpty(c.OwnerID).
			WithFieldTrace(s, &s.ClusterValues.OwnerID),
		// MeasurementSalt needs to be filled.
		validation.NotEmptySlice(c.MeasurementSalt).
			WithFieldTrace(s, &s.ClusterValues.MeasurementSalt),
	)
}

// HexBytes is a byte slice that is marshalled to and from a hex string.
type HexBytes []byte

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (h *HexBytes) UnmarshalYAML(unmarshal func(any) error) error {
	var hexString string
	if err := unmarshal(&hexString); err != nil {
		// TODO(msanft): Remove with v2.14.0
		// fall back to unmarshalling as a byte slice for backwards compatibility
		var oldHexBytes []byte
		if err := unmarshal(&oldHexBytes); err != nil {
			return fmt.Errorf("unmarshalling hex bytes: %w", err)
		}
		hexString = hex.EncodeToString(oldHexBytes)
	}

	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return fmt.Errorf("decoding hex bytes: %w", err)
	}

	*h = bytes
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
func (h HexBytes) MarshalYAML() (any, error) {
	return hex.EncodeToString(h), nil
}
