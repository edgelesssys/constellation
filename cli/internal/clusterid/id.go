/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package clusterid

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// File contains identifying information about a cluster.
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
	// InitSecret is the secret the first Bootstrapper uses to verify the user.
	InitSecret []byte `json:"initsecret,omitempty"`
	// AttestationURL is the URL of the attestation service.
	// It is only set if the cluster is created on Azure.
	AttestationURL string `json:"attestationURL,omitempty"`
}
