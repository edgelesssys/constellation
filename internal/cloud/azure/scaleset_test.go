/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScaleSetVM(t *testing.T) {
	wantInstance := metadata.InstanceMetadata{
		Name:       "scale-set-name-instance-id",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		Role:       role.Worker,
		VPCIP:      "192.0.2.0",
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
			Role:       role.Worker,
			VPCIP:      "192.0.2.0",
			SSHKeys:    map[string][]string{"user": {"key-data"}},
		},
	}
	testCases := map[string]struct {
		networkInterfacesAPI         networkInterfacesAPI
		virtualMachineScaleSetVMsAPI virtualMachineScaleSetVMsAPI
		scaleSetsAPI                 scaleSetsAPI
		wantErr                      bool
		wantInstances                []metadata.InstanceMetadata
	}{
		"listVMs works": {
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			wantInstances:                wantInstances,
		},
		"invalid scale sets are skipped": {
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newVirtualMachineScaleSetsVMsStub(),
			scaleSetsAPI:                 newListContainingNilScaleSetStub(),
			wantInstances:                wantInstances,
		},
		"listVMs can return 0 VMs": {
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: &stubVirtualMachineScaleSetVMsAPI{pager: &stubVirtualMachineScaleSetVMPager{}},
			scaleSetsAPI:                 newScaleSetsStub(),
			wantInstances:                []metadata.InstanceMetadata{},
		},
		"can skip nil in VM list": {
			networkInterfacesAPI:         newNetworkInterfacesStub(),
			virtualMachineScaleSetVMsAPI: newListContainingNilScaleSetVirtualMachinesStub(),
			scaleSetsAPI:                 newScaleSetsStub(),
			wantInstances:                wantInstances,
		},
		"converting instance fails": {
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
		inVM         armcomputev2.VirtualMachineScaleSetVM
		inInterface  []armnetwork.Interface
		inPublicIP   string
		wantErr      bool
		wantInstance metadata.InstanceMetadata
	}{
		"conversion works": {
			inVM: armcomputev2.VirtualMachineScaleSetVM{
				Name: to.Ptr("scale-set-name_instance-id"),
				ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
				Tags: map[string]*string{"tag-key": to.Ptr("tag-value")},
				Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcomputev2.OSProfile{
						ComputerName: to.Ptr("scale-set-name-instance-id"),
					},
				},
			},
			inInterface: []armnetwork.Interface{
				{
					Name: to.Ptr("scale-set-name_instance-id"),
					ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/interface-name"),
					Properties: &armnetwork.InterfacePropertiesFormat{
						IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
							{
								Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
									Primary:          to.Ptr(true),
									PrivateIPAddress: to.Ptr("192.0.2.0"),
								},
							},
						},
					},
				},
			},
			inPublicIP: "192.0.2.100",
			wantInstance: metadata.InstanceMetadata{
				Name:       "scale-set-name-instance-id",
				ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
				VPCIP:      "192.0.2.0",
				PublicIP:   "192.0.2.100",
				SSHKeys:    map[string][]string{},
			},
		},
		"invalid instance": {
			inVM:    armcomputev2.VirtualMachineScaleSetVM{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instance, err := convertScaleSetVMToCoreInstance(tc.inVM, tc.inInterface, tc.inPublicIP)

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
		tags     map[string]*string
		wantRole role.Role
	}{
		"control-plane role": {
			tags:     map[string]*string{"role": to.Ptr("control-plane")},
			wantRole: role.ControlPlane,
		},
		"worker role": {
			tags:     map[string]*string{"role": to.Ptr("worker")},
			wantRole: role.Worker,
		},
		"unknown role": {
			tags:     map[string]*string{"role": to.Ptr("foo")},
			wantRole: role.Unknown,
		},
		"no role": {
			tags:     map[string]*string{},
			wantRole: role.Unknown,
		},
		"nil role": {
			tags:     map[string]*string{"role": nil},
			wantRole: role.Unknown,
		},
		"nil tags": {
			tags:     nil,
			wantRole: role.Unknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			role := extractScaleSetVMRole(tc.tags)

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
		getVM: armcomputev2.VirtualMachineScaleSetVM{},
	}
}

func newListContainingNilScaleSetVirtualMachinesStub() *stubVirtualMachineScaleSetVMsAPI {
	return &stubVirtualMachineScaleSetVMsAPI{
		pager: &stubVirtualMachineScaleSetVMPager{
			list: []armcomputev2.VirtualMachineScaleSetVM{
				{
					Name:       to.Ptr("scale-set-name_instance-id"),
					ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
					InstanceID: to.Ptr("instance-id"),
					Tags: map[string]*string{
						"role": to.Ptr("worker"),
					},
					Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
						NetworkProfile: &armcomputev2.NetworkProfile{
							NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
								{
									ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
								},
							},
						},
						OSProfile: &armcomputev2.OSProfile{
							ComputerName: to.Ptr("scale-set-name-instance-id"),
							LinuxConfiguration: &armcomputev2.LinuxConfiguration{
								SSH: &armcomputev2.SSHConfiguration{
									PublicKeys: []*armcomputev2.SSHPublicKey{
										{
											KeyData: to.Ptr("key-data"),
											Path:    to.Ptr("/home/user/.ssh/authorized_keys"),
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
		pager: &stubVirtualMachineScaleSetVMPager{
			list: []armcomputev2.VirtualMachineScaleSetVM{
				{
					InstanceID: to.Ptr("instance-id"),
					Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
						OSProfile: &armcomputev2.OSProfile{},
						NetworkProfile: &armcomputev2.NetworkProfile{
							NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
								{
									ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id/networkInterfaces/interface-name"),
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
		pager: &stubVirtualMachineScaleSetsClientListPager{
			list: []armcomputev2.VirtualMachineScaleSet{{Name: to.Ptr("scale-set-name")}},
		},
	}
}

func TestImageReferenceFromImage(t *testing.T) {
	testCases := map[string]struct {
		img             string
		wantID          *string
		wantCommunityID *string
	}{
		"ID": {
			img:             "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.5.0",
			wantID:          to.Ptr("/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.5.0"),
			wantCommunityID: nil,
		},
		"Community": {
			img:             "/CommunityGalleries/ConstellationCVM-728bd310-e898-4450-a1ed-21cf2fb0d735/Images/feat-azure-cvm-sharing/Versions/2022.0826.084922",
			wantID:          nil,
			wantCommunityID: to.Ptr("/CommunityGalleries/ConstellationCVM-728bd310-e898-4450-a1ed-21cf2fb0d735/Images/feat-azure-cvm-sharing/Versions/2022.0826.084922"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ref := ImageReferenceFromImage(tc.img)

			assert.Equal(tc.wantID, ref.ID)
			assert.Equal(tc.wantCommunityID, ref.CommunityGalleryImageID)
		})
	}
}
