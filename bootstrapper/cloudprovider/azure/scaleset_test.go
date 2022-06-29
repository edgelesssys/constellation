package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScaleSetVM(t *testing.T) {
	wantInstance := metadata.InstanceMetadata{
		Name:       "scale-set-name-instance-id",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		PrivateIPs: []string{"192.0.2.0"},
		SSHKeys:    map[string][]string{"user": {"key-data"}},
	}
	testCases := map[string]struct {
		providerID                   string
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		wantErr                      bool
		wantInstance                 metadata.InstanceMetadata
	}{
		"getVM for scale set instance works": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			wantInstance:                 wantInstance,
		},
		"getVM for individual instance must fail": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			wantErr:    true,
		},
		"Get fails": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			virtualMachineScaleSetVMsAPI: newFailingGetScaleSetVirtualMachinesStub(),
			wantErr:                      true,
		},
		"conversion fails": {
			providerID:                   "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			virtualMachineScaleSetVMsAPI: newGetInvalidScaleSetVirtualMachinesStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			wantErr:                      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
			}
			instance, err := metadata.getScaleSetVM(context.Background(), tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestListScaleSetVMs(t *testing.T) {
	wantInstances := []metadata.InstanceMetadata{
		{
			Name:       "scale-set-name-instance-id",
			ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			PrivateIPs: []string{"192.0.2.0"},
			SSHKeys:    map[string][]string{"user": {"key-data"}},
		},
	}
	testCases := map[string]struct {
		imdsAPI                      imdsAPI
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		scaleSetsAPI                 scaleSetsAPI
		wantErr                      bool
		wantInstances                []metadata.InstanceMetadata
	}{
		"listVMs works": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			wantInstances:                wantInstances,
		},
		"invalid scale sets are skipped": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newListContainingNilScaleSetStub(),
			wantInstances:                wantInstances,
		},
		"listVMs can return 0 VMs": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: &stubVirtualMachineScaleSetVMsAPI{},
			scaleSetsAPI:                 newScaleSetsStub(),
			wantInstances:                []metadata.InstanceMetadata{},
		},
		"can skip nil in VM list": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newListContainingNilScaleSetVirtualMachinesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			wantInstances:                wantInstances,
		},
		"converting instance fails": {
			imdsAPI:                      newScaleSetIMDSStub(),
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newListContainingInvalidScaleSetVirtualMachinesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			wantErr:                      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Metadata{
				imdsAPI:                      tc.imdsAPI,
				networkInterfacesAPI:         tc.networkInterfacesAPI,
				virtualMachineScaleSetVMsAPI: tc.virtualMachineScaleSetVMsAPI,
				scaleSetsAPI:                 tc.scaleSetsAPI,
			}
			instances, err := metadata.listScaleSetVMs(context.Background(), "resource-group")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantInstances, instances)
		})
	}
}

func TestConvertScaleSetVMToCoreInstance(t *testing.T) {
	testCases := map[string]struct {
		inVM         armcompute.VirtualMachineScaleSetVM
		inInterface  []armnetwork.Interface
		inPublicIPs  []string
		wantErr      bool
		wantInstance metadata.InstanceMetadata
	}{
		"conversion works": {
			inVM: armcompute.VirtualMachineScaleSetVM{
				Name: to.StringPtr("scale-set-name_instance-id"),
				ID:   to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
				Tags: map[string]*string{"tag-key": to.StringPtr("tag-value")},
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcompute.OSProfile{
						ComputerName: to.StringPtr("scale-set-name-instance-id"),
					},
				},
			},
			inInterface: []armnetwork.Interface{
				{
					Name: to.StringPtr("scale-set-name_instance-id"),
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
			inPublicIPs: []string{"192.0.2.100", "192.0.2.101"},
			wantInstance: metadata.InstanceMetadata{
				Name:       "scale-set-name-instance-id",
				ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
				PrivateIPs: []string{"192.0.2.0"},
				PublicIPs:  []string{"192.0.2.100", "192.0.2.101"},
				SSHKeys:    map[string][]string{},
			},
		},
		"invalid instance": {
			inVM:    armcompute.VirtualMachineScaleSetVM{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instance, err := convertScaleSetVMToCoreInstance("scale-set", tc.inVM, tc.inInterface, tc.inPublicIPs)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestExtractScaleSetVMRole(t *testing.T) {
	testCases := map[string]struct {
		scaleSet string
		wantRole role.Role
	}{
		"bootstrapper role": {
			scaleSet: "constellation-scale-set-controlplanes-abcd123",
			wantRole: role.ControlPlane,
		},
		"node role": {
			scaleSet: "constellation-scale-set-workers-abcd123",
			wantRole: role.Worker,
		},
		"unknown role": {
			scaleSet: "unknown",
			wantRole: role.Unknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			role := extractScaleSetVMRole(tc.scaleSet)

			assert.Equal(tc.wantRole, role)
		})
	}
}

func newFailingGetScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		getErr: errors.New("get err"),
	}
}

func newGetInvalidScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		getVM: armcompute.VirtualMachineScaleSetVM{},
	}
}

func newListContainingNilScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSetVM{
			{
				nil,
				{
					Name:       to.StringPtr("scale-set-name_instance-id"),
					ID:         to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
					InstanceID: to.StringPtr("instance-id"),
					Tags: map[string]*string{
						"tag-key": to.StringPtr("tag-value"),
					},
					Properties: &armcompute.VirtualMachineScaleSetVMProperties{
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcompute.OSProfile{
							ComputerName: to.StringPtr("scale-set-name-instance-id"),
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

func newListContainingInvalidScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSetVM{
			{
				{
					Name:       nil,
					ID:         nil,
					InstanceID: to.StringPtr("instance-id"),
					Properties: &armcompute.VirtualMachineScaleSetVMProperties{
						OSProfile: &armcompute.OSProfile{
							ComputerName: nil,
						},
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.StringPtr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func newListContainingNilScaleSetStub() *stubScaleSetsAPI {
	return &stubScaleSetsAPI{
		listPages: [][]*armcompute.VirtualMachineScaleSet{
			{
				nil,
				{Name: to.StringPtr("scale-set-name")},
			},
		},
	}
}
