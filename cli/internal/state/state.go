/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package state defines the structure of the Constellation state file.
package state

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/config"
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
	Version        string         `yaml:"version"`
	Infrastructure Infrastructure `yaml:"infrastructure"`
	ClusterValues  ClusterValues  `yaml:"clusterValues"`
}

// New creates a new cluster state (file).
func New() *State {
	return &State{
		Version: Version1,
	}
}

// NewFromIDFile creates a new cluster state file from the given ID file.
func NewFromIDFile(idFile clusterid.File) *State {
	s := New().
		SetClusterValues(ClusterValues{
			OwnerID:         idFile.OwnerID,
			ClusterID:       idFile.ClusterID,
			MeasurementSalt: idFile.MeasurementSalt,
		}).
		SetInfrastructure(Infrastructure{
			UID:               idFile.UID,
			ClusterEndpoint:   idFile.IP,
			APIServerCertSANs: idFile.APIServerCertSANs,
			InitSecret:        string(idFile.InitSecret),
		})

	if idFile.AttestationURL != "" {
		s.Infrastructure.Azure = &Azure{
			AttestationURL: idFile.AttestationURL,
		}
	}

	return s
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

// ClusterName returns the name of the cluster.
func (s *State) ClusterName(cfg *config.Config) string {
	return cfg.Name + "-" + s.Infrastructure.UID
}

// ClusterValues describe the (Kubernetes) cluster state, set during initialization of the cluster.
type ClusterValues struct {
	// ClusterID is the unique identifier of the cluster.
	ClusterID string `yaml:"clusterID"`
	// OwnerID is the unique identifier of the owner of the cluster.
	OwnerID string `yaml:"ownerID"`
	// MeasurementSalt is the salt generated during cluster init.
	MeasurementSalt []byte `yaml:"measurementSalt"`
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
