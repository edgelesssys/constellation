/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

import (
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
)

// ConstellationState is the state of a Constellation.
type ConstellationState struct {
	Name           string `json:"name,omitempty"`
	UID            string `json:"uid,omitempty"`
	CloudProvider  string `json:"cloudprovider,omitempty"`
	LoadBalancerIP string `json:"bootstrapperhost,omitempty"`

	GCPWorkerInstances              cloudtypes.Instances `json:"gcpworkers,omitempty"`
	GCPControlPlaneInstances        cloudtypes.Instances `json:"gcpcontrolplanes,omitempty"`
	GCPWorkerInstanceGroup          string               `json:"gcpworkerinstancegroup,omitempty"`
	GCPControlPlaneInstanceGroup    string               `json:"gcpcontrolplaneinstancegroup,omitempty"`
	GCPWorkerInstanceTemplate       string               `json:"gcpworkerinstancetemplate,omitempty"`
	GCPControlPlaneInstanceTemplate string               `json:"gcpcontrolplaneinstancetemplate,omitempty"`
	GCPNetwork                      string               `json:"gcpnetwork,omitempty"`
	GCPSubnetwork                   string               `json:"gcpsubnetwork,omitempty"`
	GCPFirewalls                    []string             `json:"gcpfirewalls,omitempty"`
	GCPLoadbalancerIPname           string               `json:"gcploadbalanceripid,omitempty"`
	GCPLoadbalancers                []string             `json:"gcploadbalancers,omitempty"`
	GCPProject                      string               `json:"gcpproject,omitempty"`
	GCPZone                         string               `json:"gcpzone,omitempty"`
	GCPRegion                       string               `json:"gcpregion,omitempty"`

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
