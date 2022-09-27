/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
)

// ConstellationState is the state of a Constellation.
type ConstellationState struct {
	Name           string `json:"name,omitempty"`
	UID            string `json:"uid,omitempty"`
	CloudProvider  string `json:"cloudprovider,omitempty"`
	LoadBalancerIP string `json:"bootstrapperhost,omitempty"`

	AzureWorkerInstances       cloudtypes.Instances `json:"azureworkers,omitempty"`
	AzureControlPlaneInstances cloudtypes.Instances `json:"azurecontrolplanes,omitempty"`
	AzureResourceGroup         string               `json:"azureresourcegroup,omitempty"`
	AzureLocation              string               `json:"azurelocation,omitempty"`
	AzureSubscription          string               `json:"azuresubscription,omitempty"`
	AzureTenant                string               `json:"azuretenant,omitempty"`
	AzureSubnet                string               `json:"azuresubnet,omitempty"`
	AzureNetworkSecurityGroup  string               `json:"azurenetworksecuritygroup,omitempty"`
	AzureWorkerScaleSet        string               `json:"azureworkersscaleset,omitempty"`
	AzureControlPlaneScaleSet  string               `json:"azurecontrolplanesscaleset,omitempty"`
	AzureADAppObjectID         string               `json:"azureadappobjectid,omitempty"`

	QEMUWorkerInstances       cloudtypes.Instances `json:"qemuworkers,omitempty"`
	QEMUControlPlaneInstances cloudtypes.Instances `json:"qemucontrolplanes,omitempty"`
}
