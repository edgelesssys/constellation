package azure

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVMInterfaces(t *testing.T) {
	wantConfigs := []*armnetwork.InterfaceIPConfiguration{
		{
			Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
				PrivateIPAddress: to.StringPtr("192.0.2.0"),
			},
		},
	}
	vm := armcompute.VirtualMachine{
		Properties: &armcompute.VirtualMachineProperties{
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
					},
				},
			},
		},
	}
	testCases := map[string]struct {
		vm                   armcompute.VirtualMachine
		networkInterfacesAPI networkInterfacesAPI
		wantErr              bool
		wantConfigs          []*armnetwork.InterfaceIPConfiguration
	}{
		"retrieval works": {
			vm:                   vm,
			networkInterfacesAPI: newNetworkInterfacesStub(),
			wantConfigs:          wantConfigs,
		},
		"vm can have 0 interfaces": {
			vm:                   armcompute.VirtualMachine{},
			networkInterfacesAPI: newNetworkInterfacesStub(),
			wantConfigs:          []*armnetwork.InterfaceIPConfiguration{},
		},
		"interface retrieval fails": {
			vm:                   vm,
			networkInterfacesAPI: newFailingNetworkInterfacesStub(),
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI: tc.networkInterfacesAPI,
			}
			configs, err := metadata.getVMInterfaces(context.Background(), tc.vm, "resource-group")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantConfigs, configs)
		})
	}
}

func TestGetScaleSetVMInterfaces(t *testing.T) {
	wantConfigs := []*armnetwork.InterfaceIPConfiguration{
		{
			Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
				PrivateIPAddress: to.StringPtr("192.0.2.0"),
			},
		},
	}
	vm := armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
					},
				},
			},
		},
	}
	testCases := map[string]struct {
		vm                   armcompute.VirtualMachineScaleSetVM
		networkInterfacesAPI networkInterfacesAPI
		wantErr              bool
		wantConfigs          []*armnetwork.InterfaceIPConfiguration
	}{
		"retrieval works": {
			vm:                   vm,
			networkInterfacesAPI: newNetworkInterfacesStub(),
			wantConfigs:          wantConfigs,
		},
		"vm can have 0 interfaces": {
			vm:                   armcompute.VirtualMachineScaleSetVM{},
			networkInterfacesAPI: newNetworkInterfacesStub(),
			wantConfigs:          []*armnetwork.InterfaceIPConfiguration{},
		},
		"interface retrieval fails": {
			vm:                   vm,
			networkInterfacesAPI: newFailingNetworkInterfacesStub(),
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI: tc.networkInterfacesAPI,
			}
			configs, err := metadata.getScaleSetVMInterfaces(context.Background(), tc.vm, "resource-group", "scale-set-name", "instance-id")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantConfigs, configs)
		})
	}
}

func TestExtractPrivateIPs(t *testing.T) {
	testCases := map[string]struct {
		interfaceIPConfigs []*armnetwork.InterfaceIPConfiguration
		wantIPs            []string
	}{
		"extraction works": {
			interfaceIPConfigs: []*armnetwork.InterfaceIPConfiguration{
				{
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAddress: to.StringPtr("192.0.2.0"),
					},
				},
			},
			wantIPs: []string{"192.0.2.0"},
		},
		"can be empty": {
			interfaceIPConfigs: []*armnetwork.InterfaceIPConfiguration{},
		},
		"invalid interface is skipped": {
			interfaceIPConfigs: []*armnetwork.InterfaceIPConfiguration{
				{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ips := extractPrivateIPs(tc.interfaceIPConfigs)

			assert.ElementsMatch(tc.wantIPs, ips)
		})
	}
}

func TestExtractInterfaceNamesFromInterfaceReferences(t *testing.T) {
	testCases := map[string]struct {
		references []*armcompute.NetworkInterfaceReference
		wantNames  []string
	}{
		"extraction with individual interface reference works": {
			references: []*armcompute.NetworkInterfaceReference{
				{
					ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
				},
			},
			wantNames: []string{"interface-name"},
		},
		"extraction with scale set interface reference works": {
			references: []*armcompute.NetworkInterfaceReference{
				{
					ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
				},
			},
			wantNames: []string{"interface-name"},
		},
		"can be empty": {
			references: []*armcompute.NetworkInterfaceReference{},
		},
		"interface reference containing nil fields is skipped": {
			references: []*armcompute.NetworkInterfaceReference{
				{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			names := extractInterfaceNamesFromInterfaceReferences(tc.references)

			assert.ElementsMatch(tc.wantNames, names)
		})
	}
}
