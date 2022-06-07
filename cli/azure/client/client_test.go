package client

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGetState(t *testing.T) {
	testCases := map[string]struct {
		state   state.ConstellationState
		wantErr bool
	}{
		"valid state": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
		},
		"missing nodes": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing coordinator": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing name": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing uid": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing resource group": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing location": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing subscription": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureTenant:               "tenant",
				AzureLocation:             "location",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing tenant": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureSubscription:         "subscription",
				AzureLocation:             "location",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing subnet": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing network security group": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNodesScaleSet:        "node-scale-set",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing node scale set": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureCoordinatorsScaleSet: "coordinator-scale-set",
			},
			wantErr: true,
		},
		"missing coordinator scale set": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.Azure.String(),
				AzureNodes: azure.Instances{
					"0": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				AzureCoordinators: azure.Instances{
					"0": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				Name:                      "name",
				UID:                       "uid",
				AzureResourceGroup:        "resource-group",
				AzureLocation:             "location",
				AzureSubscription:         "subscription",
				AzureTenant:               "tenant",
				AzureSubnet:               "azure-subnet",
				AzureNetworkSecurityGroup: "network-security-group",
				AzureNodesScaleSet:        "node-scale-set",
			},
			wantErr: true,
		},
	}

	t.Run("SetState", func(t *testing.T) {
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				client := Client{}
				if tc.wantErr {
					assert.Error(client.SetState(tc.state))
				} else {
					assert.NoError(client.SetState(tc.state))
					assert.Equal(tc.state.AzureNodes, client.nodes)
					assert.Equal(tc.state.AzureCoordinators, client.coordinators)
					assert.Equal(tc.state.Name, client.name)
					assert.Equal(tc.state.UID, client.uid)
					assert.Equal(tc.state.AzureResourceGroup, client.resourceGroup)
					assert.Equal(tc.state.AzureLocation, client.location)
					assert.Equal(tc.state.AzureSubscription, client.subscriptionID)
					assert.Equal(tc.state.AzureTenant, client.tenantID)
					assert.Equal(tc.state.AzureSubnet, client.subnetID)
					assert.Equal(tc.state.AzureNetworkSecurityGroup, client.networkSecurityGroup)
					assert.Equal(tc.state.AzureNodesScaleSet, client.nodesScaleSet)
					assert.Equal(tc.state.AzureCoordinatorsScaleSet, client.coordinatorsScaleSet)
				}
			})
		}
	})

	t.Run("GetState", func(t *testing.T) {
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				client := Client{
					nodes:                tc.state.AzureNodes,
					coordinators:         tc.state.AzureCoordinators,
					name:                 tc.state.Name,
					uid:                  tc.state.UID,
					resourceGroup:        tc.state.AzureResourceGroup,
					location:             tc.state.AzureLocation,
					subscriptionID:       tc.state.AzureSubscription,
					tenantID:             tc.state.AzureTenant,
					subnetID:             tc.state.AzureSubnet,
					networkSecurityGroup: tc.state.AzureNetworkSecurityGroup,
					nodesScaleSet:        tc.state.AzureNodesScaleSet,
					coordinatorsScaleSet: tc.state.AzureCoordinatorsScaleSet,
				}
				if tc.wantErr {
					_, err := client.GetState()
					assert.Error(err)
				} else {
					state, err := client.GetState()
					assert.NoError(err)
					assert.Equal(tc.state, state)
				}
			})
		}
	})
}

func TestSetStateCloudProvider(t *testing.T) {
	assert := assert.New(t)

	client := Client{}
	stateMissingCloudProvider := state.ConstellationState{
		AzureNodes: azure.Instances{
			"0": {
				PublicIP:  "ip1",
				PrivateIP: "ip2",
			},
		},
		AzureCoordinators: azure.Instances{
			"0": {
				PublicIP:  "ip3",
				PrivateIP: "ip4",
			},
		},
		Name:                      "name",
		UID:                       "uid",
		AzureResourceGroup:        "resource-group",
		AzureLocation:             "location",
		AzureSubscription:         "subscription",
		AzureSubnet:               "azure-subnet",
		AzureNetworkSecurityGroup: "network-security-group",
		AzureNodesScaleSet:        "node-scale-set",
		AzureCoordinatorsScaleSet: "coordinator-scale-set",
	}
	assert.Error(client.SetState(stateMissingCloudProvider))
	stateIncorrectCloudProvider := state.ConstellationState{
		CloudProvider: "incorrect",
		AzureNodes: azure.Instances{
			"0": {
				PublicIP:  "ip1",
				PrivateIP: "ip2",
			},
		},
		AzureCoordinators: azure.Instances{
			"0": {
				PublicIP:  "ip3",
				PrivateIP: "ip4",
			},
		},
		Name:                      "name",
		UID:                       "uid",
		AzureResourceGroup:        "resource-group",
		AzureLocation:             "location",
		AzureSubscription:         "subscription",
		AzureSubnet:               "azure-subnet",
		AzureNetworkSecurityGroup: "network-security-group",
		AzureNodesScaleSet:        "node-scale-set",
		AzureCoordinatorsScaleSet: "coordinator-scale-set",
	}
	assert.Error(client.SetState(stateIncorrectCloudProvider))
}

func TestInit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	client := Client{}
	require.NoError(client.init("location", "name"))
	assert.Equal("location", client.location)
	assert.Equal("name", client.name)
	assert.NotEmpty(client.uid)
}
