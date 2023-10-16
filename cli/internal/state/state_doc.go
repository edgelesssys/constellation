// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Code generated by hack/docgen tool. DO NOT EDIT.

package state

import (
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
)

var (
	StateDoc          encoder.Doc
	ClusterValuesDoc  encoder.Doc
	InfrastructureDoc encoder.Doc
	GCPDoc            encoder.Doc
	AzureDoc          encoder.Doc
)

func init() {
	StateDoc.Type = "State"
	StateDoc.Comments[encoder.LineComment] = "State describe the entire state to describe a Constellation cluster."
	StateDoc.Description = "State describe the entire state to describe a Constellation cluster."
	StateDoc.Fields = make([]encoder.Doc, 3)
	StateDoc.Fields[0].Name = "version"
	StateDoc.Fields[0].Type = "string"
	StateDoc.Fields[0].Note = ""
	StateDoc.Fields[0].Description = "Schema version of this state file."
	StateDoc.Fields[0].Comments[encoder.LineComment] = "Schema version of this state file."
	StateDoc.Fields[1].Name = "infrastructure"
	StateDoc.Fields[1].Type = "Infrastructure"
	StateDoc.Fields[1].Note = ""
	StateDoc.Fields[1].Description = "State of the cluster's cloud resources. These values are retrieved during\ncluster creation. In the case of self-managed infrastructure, the marked\nfields in this struct should be filled by the user as per\nhttps://docs.edgeless.systems/constellation/workflows/create."
	StateDoc.Fields[1].Comments[encoder.LineComment] = "State of the cluster's cloud resources. These values are retrieved during"
	StateDoc.Fields[2].Name = "clusterValues"
	StateDoc.Fields[2].Type = "ClusterValues"
	StateDoc.Fields[2].Note = ""
	StateDoc.Fields[2].Description = "DO NOT EDIT. State of the Constellation Kubernetes cluster.\nThese values are set during cluster initialization and should not be changed."
	StateDoc.Fields[2].Comments[encoder.LineComment] = "DO NOT EDIT. State of the Constellation Kubernetes cluster."

	ClusterValuesDoc.Type = "ClusterValues"
	ClusterValuesDoc.Comments[encoder.LineComment] = "ClusterValues describe the (Kubernetes) cluster state, set during initialization of the cluster."
	ClusterValuesDoc.Description = "ClusterValues describe the (Kubernetes) cluster state, set during initialization of the cluster."
	ClusterValuesDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "State",
			FieldName: "clusterValues",
		},
	}
	ClusterValuesDoc.Fields = make([]encoder.Doc, 3)
	ClusterValuesDoc.Fields[0].Name = "clusterID"
	ClusterValuesDoc.Fields[0].Type = "string"
	ClusterValuesDoc.Fields[0].Note = ""
	ClusterValuesDoc.Fields[0].Description = "Unique identifier of the cluster."
	ClusterValuesDoc.Fields[0].Comments[encoder.LineComment] = "Unique identifier of the cluster."
	ClusterValuesDoc.Fields[1].Name = "ownerID"
	ClusterValuesDoc.Fields[1].Type = "string"
	ClusterValuesDoc.Fields[1].Note = ""
	ClusterValuesDoc.Fields[1].Description = "Unique identifier of the owner of the cluster."
	ClusterValuesDoc.Fields[1].Comments[encoder.LineComment] = "Unique identifier of the owner of the cluster."
	ClusterValuesDoc.Fields[2].Name = "measurementSalt"
	ClusterValuesDoc.Fields[2].Type = "hexBytes"
	ClusterValuesDoc.Fields[2].Note = ""
	ClusterValuesDoc.Fields[2].Description = "Salt used to generate the ClusterID on the bootstrapping node."
	ClusterValuesDoc.Fields[2].Comments[encoder.LineComment] = "Salt used to generate the ClusterID on the bootstrapping node."

	InfrastructureDoc.Type = "Infrastructure"
	InfrastructureDoc.Comments[encoder.LineComment] = "Infrastructure describe the state related to the cloud resources of the cluster."
	InfrastructureDoc.Description = "Infrastructure describe the state related to the cloud resources of the cluster."
	InfrastructureDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "State",
			FieldName: "infrastructure",
		},
	}
	InfrastructureDoc.Fields = make([]encoder.Doc, 7)
	InfrastructureDoc.Fields[0].Name = "uid"
	InfrastructureDoc.Fields[0].Type = "string"
	InfrastructureDoc.Fields[0].Note = ""
	InfrastructureDoc.Fields[0].Description = "Unique identifier the cluster's cloud resources are tagged with."
	InfrastructureDoc.Fields[0].Comments[encoder.LineComment] = "Unique identifier the cluster's cloud resources are tagged with."
	InfrastructureDoc.Fields[1].Name = "clusterEndpoint"
	InfrastructureDoc.Fields[1].Type = "string"
	InfrastructureDoc.Fields[1].Note = ""
	InfrastructureDoc.Fields[1].Description = "Endpoint the cluster can be reached at."
	InfrastructureDoc.Fields[1].Comments[encoder.LineComment] = "Endpoint the cluster can be reached at."
	InfrastructureDoc.Fields[2].Name = "initSecret"
	InfrastructureDoc.Fields[2].Type = "hexBytes"
	InfrastructureDoc.Fields[2].Note = ""
	InfrastructureDoc.Fields[2].Description = "Secret used to authenticate the bootstrapping node."
	InfrastructureDoc.Fields[2].Comments[encoder.LineComment] = "Secret used to authenticate the bootstrapping node."
	InfrastructureDoc.Fields[3].Name = "apiServerCertSANs"
	InfrastructureDoc.Fields[3].Type = "[]string"
	InfrastructureDoc.Fields[3].Note = ""
	InfrastructureDoc.Fields[3].Description = "description: |\n  List of Subject Alternative Names (SANs) to add to the Kubernetes API server certificate.\n If no SANs should be added, this field can be left empty.\n"
	InfrastructureDoc.Fields[3].Comments[encoder.LineComment] = "description: |"
	InfrastructureDoc.Fields[4].Name = "name"
	InfrastructureDoc.Fields[4].Type = "string"
	InfrastructureDoc.Fields[4].Note = ""
	InfrastructureDoc.Fields[4].Description = "Name used in the cluster's named resources."
	InfrastructureDoc.Fields[4].Comments[encoder.LineComment] = "Name used in the cluster's named resources."
	InfrastructureDoc.Fields[5].Name = "azure"
	InfrastructureDoc.Fields[5].Type = "Azure"
	InfrastructureDoc.Fields[5].Note = ""
	InfrastructureDoc.Fields[5].Description = "Values specific to a Constellation cluster running on Azure."
	InfrastructureDoc.Fields[5].Comments[encoder.LineComment] = "Values specific to a Constellation cluster running on Azure."
	InfrastructureDoc.Fields[6].Name = "gcp"
	InfrastructureDoc.Fields[6].Type = "GCP"
	InfrastructureDoc.Fields[6].Note = ""
	InfrastructureDoc.Fields[6].Description = "Values specific to a Constellation cluster running on GCP."
	InfrastructureDoc.Fields[6].Comments[encoder.LineComment] = "Values specific to a Constellation cluster running on GCP."

	GCPDoc.Type = "GCP"
	GCPDoc.Comments[encoder.LineComment] = "GCP describes the infra state related to GCP."
	GCPDoc.Description = "GCP describes the infra state related to GCP."
	GCPDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "Infrastructure",
			FieldName: "gcp",
		},
	}
	GCPDoc.Fields = make([]encoder.Doc, 3)
	GCPDoc.Fields[0].Name = "projectID"
	GCPDoc.Fields[0].Type = "string"
	GCPDoc.Fields[0].Note = ""
	GCPDoc.Fields[0].Description = "Project ID of the GCP project the cluster is running in."
	GCPDoc.Fields[0].Comments[encoder.LineComment] = "Project ID of the GCP project the cluster is running in."
	GCPDoc.Fields[1].Name = "ipCidrNode"
	GCPDoc.Fields[1].Type = "string"
	GCPDoc.Fields[1].Note = ""
	GCPDoc.Fields[1].Description = "CIDR range of the cluster's nodes."
	GCPDoc.Fields[1].Comments[encoder.LineComment] = "CIDR range of the cluster's nodes."
	GCPDoc.Fields[2].Name = "ipCidrPod"
	GCPDoc.Fields[2].Type = "string"
	GCPDoc.Fields[2].Note = ""
	GCPDoc.Fields[2].Description = "CIDR range of the cluster's pods."
	GCPDoc.Fields[2].Comments[encoder.LineComment] = "CIDR range of the cluster's pods."

	AzureDoc.Type = "Azure"
	AzureDoc.Comments[encoder.LineComment] = "Azure describes the infra state related to Azure."
	AzureDoc.Description = "Azure describes the infra state related to Azure."
	AzureDoc.AppearsIn = []encoder.Appearance{
		{
			TypeName:  "Infrastructure",
			FieldName: "azure",
		},
	}
	AzureDoc.Fields = make([]encoder.Doc, 6)
	AzureDoc.Fields[0].Name = "resourceGroup"
	AzureDoc.Fields[0].Type = "string"
	AzureDoc.Fields[0].Note = ""
	AzureDoc.Fields[0].Description = "Resource Group the cluster's resources are placed in."
	AzureDoc.Fields[0].Comments[encoder.LineComment] = "Resource Group the cluster's resources are placed in."
	AzureDoc.Fields[1].Name = "subscriptionID"
	AzureDoc.Fields[1].Type = "string"
	AzureDoc.Fields[1].Note = ""
	AzureDoc.Fields[1].Description = "ID of the Azure subscription the cluster is running in."
	AzureDoc.Fields[1].Comments[encoder.LineComment] = "ID of the Azure subscription the cluster is running in."
	AzureDoc.Fields[2].Name = "networkSecurityGroupName"
	AzureDoc.Fields[2].Type = "string"
	AzureDoc.Fields[2].Note = ""
	AzureDoc.Fields[2].Description = "Security group name of the cluster's resource group."
	AzureDoc.Fields[2].Comments[encoder.LineComment] = "Security group name of the cluster's resource group."
	AzureDoc.Fields[3].Name = "loadBalancerName"
	AzureDoc.Fields[3].Type = "string"
	AzureDoc.Fields[3].Note = ""
	AzureDoc.Fields[3].Description = "Name of the cluster's load balancer."
	AzureDoc.Fields[3].Comments[encoder.LineComment] = "Name of the cluster's load balancer."
	AzureDoc.Fields[4].Name = "userAssignedIdentity"
	AzureDoc.Fields[4].Type = "string"
	AzureDoc.Fields[4].Note = ""
	AzureDoc.Fields[4].Description = "ID of the UAMI the cluster's nodes are running with."
	AzureDoc.Fields[4].Comments[encoder.LineComment] = "ID of the UAMI the cluster's nodes are running with."
	AzureDoc.Fields[5].Name = "attestationURL"
	AzureDoc.Fields[5].Type = "string"
	AzureDoc.Fields[5].Note = ""
	AzureDoc.Fields[5].Description = "MAA endpoint that can be used as a fallback for veryifying the ID key digests\nin the cluster's attestation report if the enforcement policy is set accordingly.\nCan be left empty otherwise."
	AzureDoc.Fields[5].Comments[encoder.LineComment] = "MAA endpoint that can be used as a fallback for veryifying the ID key digests"
}

func (_ State) Doc() *encoder.Doc {
	return &StateDoc
}

func (_ ClusterValues) Doc() *encoder.Doc {
	return &ClusterValuesDoc
}

func (_ Infrastructure) Doc() *encoder.Doc {
	return &InfrastructureDoc
}

func (_ GCP) Doc() *encoder.Doc {
	return &GCPDoc
}

func (_ Azure) Doc() *encoder.Doc {
	return &AzureDoc
}

// GetConfigurationDoc returns documentation for the file ./state_doc.go.
func GetConfigurationDoc() *encoder.FileDoc {
	return &encoder.FileDoc{
		Name:        "Configuration",
		Description: "package state defines the structure of the Constellation state file.\n",
		Structs: []*encoder.Doc{
			&StateDoc,
			&ClusterValuesDoc,
			&InfrastructureDoc,
			&GCPDoc,
			&AzureDoc,
		},
	}
}
