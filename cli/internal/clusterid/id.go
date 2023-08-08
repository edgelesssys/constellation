/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package clusterid defines the structure of the Constellation cluster ID file.
// Logic in this package should be kept minimal.
package clusterid

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
)

// File contains state information about a cluster.
// This information is accessible after the creation
// and can be used by further operations such as initialization and upgrades.
type File struct {
	// ClusterID is the unique identifier of the cluster.
	ClusterID string `json:"clusterID,omitempty"`
	// OwnerID is the unique identifier of the owner of the cluster.
	OwnerID string `json:"ownerID,omitempty"`
	// UID is the unique identifier of the cluster, used for infrastructure management.
	UID string `json:"uid,omitempty"`
	// CloudProvider is the cloud provider of the cluster.
	CloudProvider cloudprovider.Provider `json:"cloudprovider,omitempty"`
	// IP is the IP address the cluster can be reached at (often the load balancer).
	IP string `json:"ip,omitempty"`
	// APIServerCertSANs are subject alternative names (SAN) that are added to
	// the TLS certificate of each apiserver instance.
	APIServerCertSANs []string `json:"apiServerCertSANs,omitempty"`
	// InitSecret is the secret the first Bootstrapper uses to verify the user.
	InitSecret []byte `json:"initsecret,omitempty"`
	// AttestationURL is the URL of the attestation service.
	// It is only set if the cluster is created on Azure.
	AttestationURL string `json:"attestationURL,omitempty"`
}

// Merge merges the other file into the current file and returns the result.
// If a field is set in both files, the value of the other file is used.
// This does in-place changes on the current file.
func (f *File) Merge(other File) *File {
	if other.ClusterID != "" {
		f.ClusterID = other.ClusterID
	}

	if other.OwnerID != "" {
		f.OwnerID = other.OwnerID
	}

	if other.UID != "" {
		f.UID = other.UID
	}

	if other.CloudProvider != cloudprovider.Unknown {
		f.CloudProvider = other.CloudProvider
	}

	if other.IP != "" {
		f.IP = other.IP
	}

	if other.InitSecret != nil {
		f.InitSecret = other.InitSecret
	}

	if other.AttestationURL != "" {
		f.AttestationURL = other.AttestationURL
	}

	return f
}

// GetClusterName returns the name of the cluster.
func GetClusterName(cfg *config.Config, idFile File) string {
	return cfg.Name + "-" + idFile.UID
}
