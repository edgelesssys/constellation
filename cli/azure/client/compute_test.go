package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateResourceGroup(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		resourceGroupAPI resourceGroupAPI
		wantErr          bool
	}{
		"successful create": {
			resourceGroupAPI: stubResourceGroupAPI{},
		},
		"failed create": {
			resourceGroupAPI: stubResourceGroupAPI{createErr: someErr},
			wantErr:          true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				location:         "location",
				name:             "name",
				uid:              "uid",
				resourceGroupAPI: tc.resourceGroupAPI,
				nodes:            make(azure.Instances),
				coordinators:     make(azure.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateResourceGroup(ctx))
			} else {
				assert.NoError(client.CreateResourceGroup(ctx))
				assert.Equal(client.name+"-"+client.uid, client.resourceGroup)
			}
		})
	}
}

func TestTerminateResourceGroup(t *testing.T) {
	someErr := errors.New("failed")
	clientWithResourceGroup := Client{
		resourceGroup:        "name",
		location:             "location",
		name:                 "name",
		uid:                  "uid",
		subnetID:             "subnet",
		nodesScaleSet:        "node-scale-set",
		coordinatorsScaleSet: "coordinator-scale-set",
		nodes: azure.Instances{
			"0": {
				PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1",
			},
		},
		coordinators: azure.Instances{
			"0": {
				PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1",
			},
		},
	}
	testCases := map[string]struct {
		resourceGroup    string
		resourceGroupAPI resourceGroupAPI
		client           Client
		wantErr          bool
	}{
		"successful terminate": {
			resourceGroupAPI: stubResourceGroupAPI{},
			client:           clientWithResourceGroup,
		},
		"no resource group to terminate": {
			resourceGroupAPI: stubResourceGroupAPI{},
			client:           Client{},
			resourceGroup:    "",
		},
		"failed terminate": {
			resourceGroupAPI: stubResourceGroupAPI{terminateErr: someErr},
			client:           clientWithResourceGroup,
			wantErr:          true,
		},
		"failed to poll terminate response": {
			resourceGroupAPI: stubResourceGroupAPI{stubResponse: stubResourceGroupsDeletePollerResponse{pollerErr: someErr}},
			client:           clientWithResourceGroup,
			wantErr:          true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			tc.client.resourceGroupAPI = tc.resourceGroupAPI
			ctx := context.Background()

			if tc.wantErr {
				assert.Error(tc.client.TerminateResourceGroup(ctx))
				return
			}
			assert.NoError(tc.client.TerminateResourceGroup(ctx))
			assert.Empty(tc.client.resourceGroup)
			assert.Empty(tc.client.subnetID)
			assert.Empty(tc.client.nodes)
			assert.Empty(tc.client.coordinators)
			assert.Empty(tc.client.nodesScaleSet)
			assert.Empty(tc.client.coordinatorsScaleSet)
		})
	}
}

func TestCreateInstances(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		publicIPAddressesAPI publicIPAddressesAPI
		networkInterfacesAPI networkInterfacesAPI
		scaleSetsAPI         scaleSetsAPI
		resourceGroupAPI     resourceGroupAPI
		roleAssignmentsAPI   roleAssignmentsAPI
		createInstancesInput CreateInstancesInput
		wantErr              bool
	}{
		"successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI: stubScaleSetsAPI{
				stubResponse: stubVirtualMachineScaleSetsCreateOrUpdatePollerResponse{
					pollResponse: armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse{
						VirtualMachineScaleSetsClientCreateOrUpdateResult: armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResult{
							VirtualMachineScaleSet: armcompute.VirtualMachineScaleSet{Identity: &armcompute.VirtualMachineScaleSetIdentity{PrincipalID: to.StringPtr("principal-id")}},
						},
					},
				},
			},
			resourceGroupAPI:   newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators:    3,
				CountNodes:           3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
			},
		},
		"error when creating scale set": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI:         stubScaleSetsAPI{createErr: someErr},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators:    3,
				CountNodes:           3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
			},
			wantErr: true,
		},
		"error when polling create scale set response": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI:         stubScaleSetsAPI{stubResponse: stubVirtualMachineScaleSetsCreateOrUpdatePollerResponse{pollErr: someErr}},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators:    3,
				CountNodes:           3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
			},
			wantErr: true,
		},
		"error when retrieving private IPs": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{getErr: someErr},
			scaleSetsAPI:         stubScaleSetsAPI{},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountNodes:           3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				location:             "location",
				name:                 "name",
				uid:                  "uid",
				resourceGroup:        "name",
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
				networkInterfacesAPI: tc.networkInterfacesAPI,
				scaleSetsAPI:         tc.scaleSetsAPI,
				resourceGroupAPI:     tc.resourceGroupAPI,
				roleAssignmentsAPI:   tc.roleAssignmentsAPI,
				nodes:                make(azure.Instances),
				coordinators:         make(azure.Instances),
				loadBalancerPubIP:    "lbip",
			}

			if tc.wantErr {
				assert.Error(client.CreateInstances(ctx, tc.createInstancesInput))
			} else {
				assert.NoError(client.CreateInstances(ctx, tc.createInstancesInput))
				assert.Equal(tc.createInstancesInput.CountCoordinators, len(client.coordinators))
				assert.Equal(tc.createInstancesInput.CountNodes, len(client.nodes))
				assert.NotEmpty(client.nodes["0"].PrivateIP)
				assert.NotEmpty(client.nodes["0"].PublicIP)
				assert.NotEmpty(client.coordinators["0"].PrivateIP)
				assert.Equal("lbip", client.coordinators["0"].PublicIP)
			}
		})
	}
}

// TODO: deprecate as soon as scale sets are available.
func TestCreateInstancesVMs(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		publicIPAddressesAPI publicIPAddressesAPI
		networkInterfacesAPI networkInterfacesAPI
		virtualMachinesAPI   virtualMachinesAPI
		resourceGroupAPI     resourceGroupAPI
		roleAssignmentsAPI   roleAssignmentsAPI
		createInstancesInput CreateInstancesInput
		wantErr              bool
	}{
		"successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			virtualMachinesAPI: stubVirtualMachinesAPI{
				stubResponse: stubVirtualMachinesClientCreateOrUpdatePollerResponse{
					pollResponse: armcompute.VirtualMachinesClientCreateOrUpdateResponse{VirtualMachinesClientCreateOrUpdateResult: armcompute.VirtualMachinesClientCreateOrUpdateResult{
						VirtualMachine: armcompute.VirtualMachine{
							Identity: &armcompute.VirtualMachineIdentity{PrincipalID: to.StringPtr("principal-id")},
						},
					}},
				},
			},
			resourceGroupAPI:   newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators: 3,
				CountNodes:        3,
				InstanceType:      "type",
				Image:             "image",
			},
		},
		"error when creating scale set": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			virtualMachinesAPI:   stubVirtualMachinesAPI{createErr: someErr},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators: 3,
				CountNodes:        3,
				InstanceType:      "type",
				Image:             "image",
			},
			wantErr: true,
		},
		"error when polling create scale set response": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			virtualMachinesAPI:   stubVirtualMachinesAPI{stubResponse: stubVirtualMachinesClientCreateOrUpdatePollerResponse{pollErr: someErr}},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators: 3,
				CountNodes:        3,
				InstanceType:      "type",
				Image:             "image",
			},
			wantErr: true,
		},
		"error when creating NIC": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{createErr: someErr},
			virtualMachinesAPI:   stubVirtualMachinesAPI{},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators: 3,
				CountNodes:        3,
				InstanceType:      "type",
				Image:             "image",
			},
			wantErr: true,
		},
		"error when creating public IP": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{createErr: someErr},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			virtualMachinesAPI:   stubVirtualMachinesAPI{},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators: 3,
				CountNodes:        3,
				InstanceType:      "type",
				Image:             "image",
			},
			wantErr: true,
		},
		"error when retrieving public IP": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{getErr: someErr},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			virtualMachinesAPI:   stubVirtualMachinesAPI{},
			resourceGroupAPI:     newSuccessfulResourceGroupStub(),
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountCoordinators: 3,
				CountNodes:        3,
				InstanceType:      "type",
				Image:             "image",
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			ctx := context.Background()

			client := Client{
				location:             "location",
				name:                 "name",
				uid:                  "uid",
				resourceGroup:        "name",
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
				networkInterfacesAPI: tc.networkInterfacesAPI,
				virtualMachinesAPI:   tc.virtualMachinesAPI,
				resourceGroupAPI:     tc.resourceGroupAPI,
				roleAssignmentsAPI:   tc.roleAssignmentsAPI,
				nodes:                make(azure.Instances),
				coordinators:         make(azure.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateInstancesVMs(ctx, tc.createInstancesInput))
				return
			}

			require.NoError(client.CreateInstancesVMs(ctx, tc.createInstancesInput))
			assert.Equal(tc.createInstancesInput.CountCoordinators, len(client.coordinators))
			assert.Equal(tc.createInstancesInput.CountNodes, len(client.nodes))
			assert.NotEmpty(client.nodes["0"].PrivateIP)
			assert.NotEmpty(client.nodes["0"].PublicIP)
			assert.NotEmpty(client.coordinators["0"].PrivateIP)
			assert.NotEmpty(client.coordinators["0"].PublicIP)
		})
	}
}

func newSuccessfulResourceGroupStub() *stubResourceGroupAPI {
	return &stubResourceGroupAPI{
		getResourceGroup: armresources.ResourceGroup{
			ID: to.StringPtr("resource-group-id"),
		},
	}
}
