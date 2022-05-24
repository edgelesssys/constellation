package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVM(t *testing.T) {
	wantInstance := cloudtypes.Instance{
		Name:       "instance-name",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
		PrivateIPs: []string{"192.0.2.0"},
		SSHKeys:    map[string][]string{"user": {"key-data"}},
	}
	testCases := map[string]struct {
		providerID           string
		networkInterfacesAPI networkInterfacesAPI
		virtualMachinesAPI   virtualMachinesAPI
		wantErr              bool
		wantInstance         cloudtypes.Instance
	}{
		"getVM for individual instance works": {
			providerID:           "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			networkInterfacesAPI: newNetworkInterfacesStub(),
			virtualMachinesAPI:   newVirtualMachinesStub(),
			wantInstance:         wantInstance,
		},
		"getVM for scale set instance must fail": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			wantErr:    true,
		},
		"Get fails": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			virtualMachinesAPI: newFailingGetVirtualMachinesStub(),
			wantErr:            true,
		},
		"retrieving interfaces fails": {
			providerID:           "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			virtualMachinesAPI:   newVirtualMachinesStub(),
			networkInterfacesAPI: newFailingNetworkInterfacesStub(),
			wantErr:              true,
		},
		"conversion fails": {
			providerID:           "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			virtualMachinesAPI:   newGetInvalidVirtualMachinesStub(),
			networkInterfacesAPI: newNetworkInterfacesStub(),
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI: tc.networkInterfacesAPI,
				virtualMachinesAPI:   tc.virtualMachinesAPI,
			}
			instance, err := metadata.getVM(context.Background(), tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestListVMs(t *testing.T) {
	wantInstances := []cloudtypes.Instance{
		{
			Name:       "instance-name",
			ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			PrivateIPs: []string{"192.0.2.0"},
			SSHKeys:    map[string][]string{"user": {"key-data"}},
		},
	}
	testCases := map[string]struct {
		imdsAPI              imdsAPI
		networkInterfacesAPI networkInterfacesAPI
		virtualMachinesAPI   virtualMachinesAPI
		wantErr              bool
		wantInstances        []cloudtypes.Instance
	}{
		"listVMs works": {
			imdsAPI:              newIMDSStub(),
			networkInterfacesAPI: newNetworkInterfacesStub(),
			virtualMachinesAPI:   newVirtualMachinesStub(),
			wantInstances:        wantInstances,
		},
		"listVMs can return 0 VMs": {
			imdsAPI:              newIMDSStub(),
			networkInterfacesAPI: newNetworkInterfacesStub(),
			virtualMachinesAPI:   &stubVirtualMachinesAPI{},
			wantInstances:        []cloudtypes.Instance{},
		},
		"can skip nil in VM list": {
			imdsAPI:              newIMDSStub(),
			networkInterfacesAPI: newNetworkInterfacesStub(),
			virtualMachinesAPI:   newListContainingNilVirtualMachinesStub(),
			wantInstances:        wantInstances,
		},
		"retrieving network interfaces fails": {
			imdsAPI:              newIMDSStub(),
			networkInterfacesAPI: newFailingNetworkInterfacesStub(),
			virtualMachinesAPI:   newVirtualMachinesStub(),
			wantErr:              true,
		},
		"converting instance fails": {
			imdsAPI:              newIMDSStub(),
			networkInterfacesAPI: newNetworkInterfacesStub(),
			virtualMachinesAPI:   newListContainingInvalidVirtualMachinesStub(),
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:              tc.imdsAPI,
				networkInterfacesAPI: tc.networkInterfacesAPI,
				virtualMachinesAPI:   tc.virtualMachinesAPI,
			}
			instances, err := metadata.listVMs(context.Background(), "resource-group")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantInstances, instances)
		})
	}
}

func TestSetTag(t *testing.T) {
	testCases := map[string]struct {
		imdsAPI imdsAPI
		tagsAPI tagsAPI
		wantErr bool
	}{
		"setTag works": {
			imdsAPI: newIMDSStub(),
			tagsAPI: newTagsStub(),
		},
		"retrieving resource ID fails": {
			imdsAPI: newFailingIMDSStub(),
			tagsAPI: newTagsStub(),
			wantErr: true,
		},
		"updating tags fails": {
			imdsAPI: newIMDSStub(),
			tagsAPI: newFailingTagsStub(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI: tc.imdsAPI,
				tagsAPI: tc.tagsAPI,
			}
			err := metadata.setTag(context.Background(), "key", "value")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestSplitVMProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID         string
		wantErr            bool
		wantSubscriptionID string
		wantResourceGroup  string
		wantInstanceName   string
	}{
		"providerID for individual instance works": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			wantSubscriptionID: "subscription-id",
			wantResourceGroup:  "resource-group",
			wantInstanceName:   "instance-name",
		},
		"providerID for scale set instance must fail": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			wantErr:    true,
		},
		"providerID is malformed": {
			providerID: "malformed-provider-id",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			subscriptionID, resourceGroup, instanceName, err := splitVMProviderID(tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantSubscriptionID, subscriptionID)
			assert.Equal(tc.wantResourceGroup, resourceGroup)
			assert.Equal(tc.wantInstanceName, instanceName)
		})
	}
}

func TestConvertVMToCoreInstance(t *testing.T) {
	testCases := map[string]struct {
		inVM         armcompute.VirtualMachine
		inInterface  []armnetwork.Interface
		wantErr      bool
		wantInstance cloudtypes.Instance
	}{
		"conversion works": {
			inVM: armcompute.VirtualMachine{
				Name: to.StringPtr("instance-name"),
				ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"),
				Tags: map[string]*string{"tag-key": to.StringPtr("tag-value")},
				Properties: &armcompute.VirtualMachineProperties{
					OSProfile: &armcompute.OSProfile{
						LinuxConfiguration: &armcompute.LinuxConfiguration{
							SSH: &armcompute.SSHConfiguration{
								PublicKeys: []*armcompute.SSHPublicKey{
									{
										Path:    to.StringPtr("/home/user/.ssh/authorized_keys"),
										KeyData: to.StringPtr("key-data"),
									},
								},
							},
						},
					},
				},
			},
			inInterface: []armnetwork.Interface{
				{
					Name: to.StringPtr("interface-name"),
					ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
					Properties: &armnetwork.InterfacePropertiesFormat{
						IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
							{
								Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
									PrivateIPAddress: to.StringPtr("192.0.2.0"),
								},
							},
						},
					},
				},
			},
			wantInstance: cloudtypes.Instance{
				Name:       "instance-name",
				ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
				PrivateIPs: []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{"user": {"key-data"}},
			},
		},
		"conversion without SSH keys works": {
			inVM: armcompute.VirtualMachine{
				Name: to.StringPtr("instance-name"),
				ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"),
				Tags: map[string]*string{"tag-key": to.StringPtr("tag-value")},
			},
			inInterface: []armnetwork.Interface{
				{
					Name: to.StringPtr("interface-name"),
					ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
					Properties: &armnetwork.InterfacePropertiesFormat{
						IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
							{
								Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
									PrivateIPAddress: to.StringPtr("192.0.2.0"),
								},
							},
						},
					},
				},
			},
			wantInstance: cloudtypes.Instance{
				Name:       "instance-name",
				ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
				PrivateIPs: []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"invalid instance": {
			inVM:    armcompute.VirtualMachine{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instance, err := convertVMToCoreInstance(tc.inVM, tc.inInterface)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func newFailingNetworkInterfacesStub() *stubNetworkInterfacesAPI {
	return &stubNetworkInterfacesAPI{
		getErr: errors.New("get err"),
	}
}

func newGetInvalidVirtualMachinesStub() *stubVirtualMachinesAPI {
	return &stubVirtualMachinesAPI{
		getVM: armcompute.VirtualMachine{},
	}
}

func newListContainingNilVirtualMachinesStub() *stubVirtualMachinesAPI {
	return &stubVirtualMachinesAPI{
		listPages: [][]*armcompute.VirtualMachine{
			{
				nil,
				{
					Name: to.StringPtr("instance-name"),
					ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"),
					Tags: map[string]*string{
						"tag-key": to.StringPtr("tag-value"),
					},
					Properties: &armcompute.VirtualMachineProperties{
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcompute.OSProfile{
							LinuxConfiguration: &armcompute.LinuxConfiguration{
								SSH: &armcompute.SSHConfiguration{
									PublicKeys: []*armcompute.SSHPublicKey{
										{
											KeyData: to.StringPtr("key-data"),
											Path:    to.StringPtr("/home/user/.ssh/authorized_keys"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newListContainingInvalidVirtualMachinesStub() *stubVirtualMachinesAPI {
	return &stubVirtualMachinesAPI{
		listPages: [][]*armcompute.VirtualMachine{
			{
				{
					Name: nil,
					ID:   nil,
					Properties: &armcompute.VirtualMachineProperties{
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcompute.OSProfile{
							LinuxConfiguration: &armcompute.LinuxConfiguration{
								SSH: &armcompute.SSHConfiguration{
									PublicKeys: []*armcompute.SSHPublicKey{
										{
											KeyData: to.StringPtr("key-data"),
											Path:    to.StringPtr("/home/user/.ssh/authorized_keys"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
