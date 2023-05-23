/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package iamid contains the output information of IAM resource creation.
*/
package iamid

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// File contains output information of an IAM configuration.
type File struct {
	// CloudProvider is the cloud provider of the cluster.
	CloudProvider cloudprovider.Provider `json:"cloudprovider,omitempty"`

	GCPOutput GCPFile `json:"gcpOutput,omitempty"`

	AzureOutput AzureFile `json:"azureOutput,omitempty"`

	AWSOutput AWSFile `json:"awsOutput,omitempty"`
}

// GCPFile contains the output information of a GCP IAM configuration.
type GCPFile struct {
	ServiceAccountKey string `json:"serviceAccountID,omitempty"`
}

// AzureFile contains the output information of a Microsoft Azure IAM configuration.
type AzureFile struct {
	SubscriptionID string `json:"subscriptionID,omitempty"`
	TenantID       string `json:"tenantID,omitempty"`
	UAMIID         string `json:"uamiID,omitempty"`
}

// AWSFile contains the output information of an AWS IAM configuration.
type AWSFile struct {
	ControlPlaneInstanceProfile string `json:"controlPlaneInstanceProfile,omitempty"`
	WorkerNodeInstanceProfile   string `json:"workerNodeInstanceProfile,omitempty"`
}
