/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
)

const (
	graphAPIResource      = "https://graph.windows.net"
	managementAPIResource = "https://management.azure.com"
)

// Client is a client for Azure.
type Client struct {
	networksAPI
	networkSecurityGroupsAPI
	resourceAPI
	scaleSetsAPI
	publicIPAddressesAPI
	networkInterfacesAPI
	loadBalancersAPI
	applicationsAPI
	servicePrincipalsAPI
	roleAssignmentsAPI
	applicationInsightsAPI

	pollFrequency time.Duration

	workers       cloudtypes.Instances
	controlPlanes cloudtypes.Instances

	name                 string
	uid                  string
	resourceGroup        string
	location             string
	subscriptionID       string
	tenantID             string
	subnetID             string
	controlPlaneScaleSet string
	workerScaleSet       string
	loadBalancerName     string
	loadBalancerPubIP    string
	networkSecurityGroup string
	adAppObjectID        string
}

// NewFromDefault creates a client with initialized clients.
func NewFromDefault(subscriptionID, tenantID string) (*Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	graphAuthorizer, err := getAuthorizer(graphAPIResource)
	if err != nil {
		return nil, err
	}
	managementAuthorizer, err := getAuthorizer(managementAPIResource)
	if err != nil {
		return nil, err
	}
	netAPI, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	netSecGrpAPI, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	scaleSetAPI, err := armcomputev2.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	publicIPAddressesAPI, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	networkInterfacesAPI, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	loadBalancersAPI, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	applicationInsightsAPI, err := armapplicationinsights.NewComponentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	resourceAPI, err := armresources.NewClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	applicationsAPI := graphrbac.NewApplicationsClient(tenantID)
	applicationsAPI.Authorizer = graphAuthorizer
	servicePrincipalsAPI := graphrbac.NewServicePrincipalsClient(tenantID)
	servicePrincipalsAPI.Authorizer = graphAuthorizer
	roleAssignmentsAPI := authorization.NewRoleAssignmentsClient(subscriptionID)
	roleAssignmentsAPI.Authorizer = managementAuthorizer

	return &Client{
		networksAPI:              netAPI,
		networkSecurityGroupsAPI: netSecGrpAPI,
		resourceAPI:              resourceAPI,
		scaleSetsAPI:             scaleSetAPI,
		publicIPAddressesAPI:     publicIPAddressesAPI,
		networkInterfacesAPI:     networkInterfacesAPI,
		loadBalancersAPI:         loadBalancersAPI,
		applicationsAPI:          applicationsAPI,
		servicePrincipalsAPI:     servicePrincipalsAPI,
		roleAssignmentsAPI:       roleAssignmentsAPI,
		applicationInsightsAPI:   applicationInsightsAPI,
		subscriptionID:           subscriptionID,
		tenantID:                 tenantID,
		workers:                  cloudtypes.Instances{},
		controlPlanes:            cloudtypes.Instances{},
		pollFrequency:            time.Second * 5,
	}, nil
}

// NewInitialized creates and initializes client by setting the subscriptionID, location and name
// of the Constellation.
func NewInitialized(subscriptionID, tenantID, name, location, resourceGroup string) (*Client, error) {
	client, err := NewFromDefault(subscriptionID, tenantID)
	if err != nil {
		return nil, err
	}
	err = client.init(location, name, resourceGroup)
	return client, err
}

// init initializes the client.
func (c *Client) init(location, name, resourceGroup string) error {
	c.location = location
	c.name = name
	c.resourceGroup = resourceGroup
	uid, err := c.generateUID()
	if err != nil {
		return err
	}
	c.uid = uid

	return nil
}

// GetState returns the state of the client as ConstellationState.
func (c *Client) GetState() state.ConstellationState {
	return state.ConstellationState{
		Name:                       c.name,
		UID:                        c.uid,
		CloudProvider:              cloudprovider.Azure.String(),
		LoadBalancerIP:             c.loadBalancerPubIP,
		AzureLocation:              c.location,
		AzureSubscription:          c.subscriptionID,
		AzureTenant:                c.tenantID,
		AzureResourceGroup:         c.resourceGroup,
		AzureNetworkSecurityGroup:  c.networkSecurityGroup,
		AzureSubnet:                c.subnetID,
		AzureWorkerScaleSet:        c.workerScaleSet,
		AzureControlPlaneScaleSet:  c.controlPlaneScaleSet,
		AzureWorkerInstances:       c.workers,
		AzureControlPlaneInstances: c.controlPlanes,
		AzureADAppObjectID:         c.adAppObjectID,
	}
}

// SetState sets the state of the client to the handed ConstellationState.
func (c *Client) SetState(stat state.ConstellationState) {
	c.resourceGroup = stat.AzureResourceGroup
	c.name = stat.Name
	c.uid = stat.UID
	c.loadBalancerPubIP = stat.LoadBalancerIP
	c.location = stat.AzureLocation
	c.subscriptionID = stat.AzureSubscription
	c.tenantID = stat.AzureTenant
	c.subnetID = stat.AzureSubnet
	c.networkSecurityGroup = stat.AzureNetworkSecurityGroup
	c.workerScaleSet = stat.AzureWorkerScaleSet
	c.controlPlaneScaleSet = stat.AzureControlPlaneScaleSet
	c.workers = stat.AzureWorkerInstances
	c.controlPlanes = stat.AzureControlPlaneInstances
	c.adAppObjectID = stat.AzureADAppObjectID
}

func (c *Client) generateUID() (string, error) {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	const uidLen = 5
	uid := make([]byte, uidLen)
	for i := 0; i < uidLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		uid[i] = letters[n.Int64()]
	}
	return string(uid), nil
}

// getAuthorizer creates an autorest.Authorizer for different Azure AD APIs using either environment variables or azure cli credentials.
func getAuthorizer(resource string) (autorest.Authorizer, error) {
	authorizer, cliErr := auth.NewAuthorizerFromCLIWithResource(resource)
	if cliErr == nil {
		return authorizer, nil
	}
	authorizer, envErr := auth.NewAuthorizerFromEnvironmentWithResource(resource)
	if envErr == nil {
		return authorizer, nil
	}
	return nil, fmt.Errorf("unable to create authorizer from env or cli: %v %v", envErr, cliErr)
}
