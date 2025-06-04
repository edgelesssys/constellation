/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestGetInstance(t *testing.T) {
	someErr := errors.New("failed")
	sampleProviderID := "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"
	successVMAPI := &stubVirtualMachineScaleSetVMsAPI{
		getVM: armcompute.VirtualMachineScaleSetVM{
			Name: to.Ptr("scale-set-name-instance-id"),
			ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
			Properties: &armcompute.VirtualMachineScaleSetVMProperties{
				OSProfile: &armcompute.OSProfile{
					ComputerName: to.Ptr("scale-set-name-instance-id"),
				},
				NetworkProfile: &armcompute.NetworkProfile{
					NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
						{
							ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/nic-name"),
						},
					},
				},
			},
			Tags: map[string]*string{
				cloud.TagRole: to.Ptr(role.Worker.String()),
			},
		},
	}
	successNetworkAPI := &stubNetworkInterfacesAPI{
		getInterface: armnetwork.Interface{
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							Primary:          to.Ptr(true),
							PrivateIPAddress: to.Ptr("192.0.2.1"),
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		scaleSetsVMAPI       *stubVirtualMachineScaleSetVMsAPI
		networkInterfacesAPI *stubNetworkInterfacesAPI
		IMDSAPI              *stubIMDSAPI
		providerID           string
		wantInstance         metadata.InstanceMetadata
		wantErr              bool
	}{
		"success worker": {
			scaleSetsVMAPI:       successVMAPI,
			networkInterfacesAPI: successNetworkAPI,
			providerID:           sampleProviderID,
			IMDSAPI:              &stubIMDSAPI{},
			wantInstance: metadata.InstanceMetadata{
				Name:       "scale-set-name-instance-id",
				ProviderID: sampleProviderID,
				Role:       role.Worker,
				VPCIP:      "192.0.2.1",
			},
		},
		"success control-plane": {
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				getVM: armcompute.VirtualMachineScaleSetVM{
					Name: to.Ptr("scale-set-name-instance-id"),
					ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
					Properties: &armcompute.VirtualMachineScaleSetVMProperties{
						OSProfile: &armcompute.OSProfile{
							ComputerName: to.Ptr("scale-set-name-instance-id"),
						},
						NetworkProfile: &armcompute.NetworkProfile{
							NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
								{
									ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Network/networkInterfaces/nic-name"),
								},
							},
						},
					},
					Tags: map[string]*string{
						cloud.TagRole: to.Ptr(role.ControlPlane.String()),
					},
				},
			},
			IMDSAPI:              &stubIMDSAPI{},
			networkInterfacesAPI: successNetworkAPI,
			providerID:           sampleProviderID,
			wantInstance: metadata.InstanceMetadata{
				Name:       "scale-set-name-instance-id",
				ProviderID: sampleProviderID,
				Role:       role.ControlPlane,
				VPCIP:      "192.0.2.1",
			},
		},
		"invalid provider ID": {
			scaleSetsVMAPI:       successVMAPI,
			IMDSAPI:              &stubIMDSAPI{},
			networkInterfacesAPI: successNetworkAPI,
			providerID:           "invalid",
			wantErr:              true,
		},
		"vm API error": {
			scaleSetsVMAPI:       &stubVirtualMachineScaleSetVMsAPI{getErr: someErr},
			IMDSAPI:              &stubIMDSAPI{},
			networkInterfacesAPI: successNetworkAPI,
			providerID:           sampleProviderID,
			wantErr:              true,
		},
		"network API error": {
			scaleSetsVMAPI:       successVMAPI,
			networkInterfacesAPI: &stubNetworkInterfacesAPI{getErr: someErr},
			IMDSAPI:              &stubIMDSAPI{},
			providerID:           sampleProviderID,
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			metadata := Cloud{
				imds:           tc.IMDSAPI,
				scaleSetsVMAPI: tc.scaleSetsVMAPI,
				netIfacAPI:     tc.networkInterfacesAPI,
			}
			instance, err := metadata.getInstance(t.Context(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestUID(t *testing.T) {
	testCases := map[string]struct {
		imdsAPI *stubIMDSAPI
		wantErr bool
	}{
		"success": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
		},
		"error": {
			imdsAPI: &stubIMDSAPI{
				uidErr: errors.New("failed"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds: tc.imdsAPI,
			}
			uid, err := cloud.UID(t.Context())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.imdsAPI.uidVal, uid)
		})
	}
}

func TestInitSecretHash(t *testing.T) {
	testCases := map[string]struct {
		imdsAPI *stubIMDSAPI
		wantErr bool
	}{
		"success": {
			imdsAPI: &stubIMDSAPI{
				initSecretHashVal: "initSecretHash",
			},
		},
		"error": {
			imdsAPI: &stubIMDSAPI{
				initSecretHashErr: errors.New("failed"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds: tc.imdsAPI,
			}
			initSecretHash, err := cloud.InitSecretHash(t.Context())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal([]byte(tc.imdsAPI.initSecretHashVal), initSecretHash)
		})
	}
}

func TestList(t *testing.T) {
	someErr := errors.New("failed")
	networkIfaceResponse := &stubNetworkInterfacesAPI{
		getInterface: armnetwork.Interface{
			Name: to.Ptr("interface-name"),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: to.Ptr("192.0.2.0"),
							Primary:          to.Ptr(true),
						},
					},
				},
			},
		},
	}
	scaleSetsResponse := &stubScaleSetsAPI{
		pager: &stubVirtualMachineScaleSetsClientListPager{
			list: []armcompute.VirtualMachineScaleSet{{
				Name: to.Ptr("scale-set"),
				Tags: map[string]*string{
					cloud.TagUID:  to.Ptr("uid"),
					cloud.TagRole: to.Ptr("worker"),
				},
			}},
		},
	}
	scaleSetVM := armcompute.VirtualMachineScaleSetVM{
		Name:       to.Ptr("scale-set_0"),
		InstanceID: to.Ptr("instance-id"),
		ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0"),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0/networkInterfaces/interface-name"),
					},
				},
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName: to.Ptr("scale-set-0"),
			},
		},
		Tags: map[string]*string{
			cloud.TagUID:  to.Ptr("uid"),
			cloud.TagRole: to.Ptr("worker"),
		},
	}

	workerInstance := metadata.InstanceMetadata{
		Name:       "scale-set-0",
		ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
		Role:       role.Worker,
		VPCIP:      "192.0.2.0",
	}

	testCases := map[string]struct {
		imdsAPI              *stubIMDSAPI
		networkInterfacesAPI *stubNetworkInterfacesAPI
		scaleSetsAPI         *stubScaleSetsAPI
		scaleSetsVMAPI       *stubVirtualMachineScaleSetVMsAPI
		wantErr              bool
		wantInstances        []metadata.InstanceMetadata
	}{
		"list single instance": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			networkInterfacesAPI: networkIfaceResponse,
			scaleSetsAPI:         scaleSetsResponse,
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				pager: &stubVirtualMachineScaleSetVMPager{
					list: []armcompute.VirtualMachineScaleSetVM{scaleSetVM},
				},
			},
			wantInstances: []metadata.InstanceMetadata{workerInstance},
		},
		"list multiple instances": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			networkInterfacesAPI: networkIfaceResponse,
			scaleSetsAPI:         scaleSetsResponse,
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				pager: &stubVirtualMachineScaleSetVMPager{
					list: []armcompute.VirtualMachineScaleSetVM{
						scaleSetVM,
						{
							Name:       to.Ptr("control-set_0"),
							InstanceID: to.Ptr("0"),
							ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/control-set/virtualMachines/0"),
							Properties: &armcompute.VirtualMachineScaleSetVMProperties{
								NetworkProfile: &armcompute.NetworkProfile{
									NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
										{
											ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/control-set/virtualMachines/0/networkInterfaces/interface-name"),
										},
									},
								},
								OSProfile: &armcompute.OSProfile{
									ComputerName: to.Ptr("control-set-0"),
								},
							},
							Tags: map[string]*string{
								cloud.TagUID:  to.Ptr("uid"),
								cloud.TagRole: to.Ptr("control-plane"),
							},
						},
					},
				},
			},
			wantInstances: []metadata.InstanceMetadata{
				workerInstance,
				{
					Name:       "control-set-0",
					ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/control-set/virtualMachines/0",
					Role:       role.ControlPlane,
					VPCIP:      "192.0.2.0",
				},
			},
		},
		"imds resource group fails": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupErr: someErr,
				uidVal:           "uid",
			},
			networkInterfacesAPI: networkIfaceResponse,
			scaleSetsAPI:         scaleSetsResponse,
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				pager: &stubVirtualMachineScaleSetVMPager{
					list: []armcompute.VirtualMachineScaleSetVM{scaleSetVM},
				},
			},
			wantErr: true,
		},
		"imds uid fails": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidErr:           someErr,
			},
			networkInterfacesAPI: networkIfaceResponse,
			scaleSetsAPI:         scaleSetsResponse,
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				pager: &stubVirtualMachineScaleSetVMPager{
					list: []armcompute.VirtualMachineScaleSetVM{scaleSetVM},
				},
			},
			wantErr: true,
		},
		"listScaleSetVMs fails": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			networkInterfacesAPI: networkIfaceResponse,
			scaleSetsAPI:         scaleSetsResponse,
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				pager: &stubVirtualMachineScaleSetVMPager{fetchErr: someErr},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			azureMetadata := Cloud{
				imds:           tc.imdsAPI,
				netIfacAPI:     tc.networkInterfacesAPI,
				scaleSetsAPI:   tc.scaleSetsAPI,
				scaleSetsVMAPI: tc.scaleSetsVMAPI,
			}
			instances, err := azureMetadata.List(t.Context())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantInstances, instances)
		})
	}
}

func TestGetNetworkSecurityGroupName(t *testing.T) {
	name := "network-security-group-name"
	testCases := map[string]struct {
		securityGroupsAPI securityGroupsAPI
		wantName          string
		wantErr           bool
	}{
		"GetNetworkSecurityGroupName works": {
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{
						{
							Tags: map[string]*string{
								cloud.TagUID: to.Ptr("uid"),
							},
							Name: to.Ptr(name),
						},
					},
				},
			},
			wantName: name,
		},
		"no security group": {
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{},
			},
			wantErr: true,
		},
		"security group API error": {
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{fetchErr: errors.New("failed")},
			},
			wantErr: true,
		},
		"missing name in security group struct": {
			securityGroupsAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{{}},
				},
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Cloud{
				secGroupAPI: tc.securityGroupsAPI,
			}
			name, err := metadata.getNetworkSecurityGroupName(t.Context(), "resource-group", "uid")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantName, name)
		})
	}
}

func TestGetSubnetworkCIDR(t *testing.T) {
	subnetworkCIDR := "192.0.2.0/24"
	name := "name"
	testCases := map[string]struct {
		virtualNetworksAPI virtualNetworksAPI
		imdsAPI            imdsAPI
		wantNetworkCIDR    string
		wantErr            bool
	}{
		"GetSubnetworkCIDR works": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			virtualNetworksAPI: &stubVirtualNetworksAPI{
				pager: &stubVirtualNetworksClientListPager{
					list: []armnetwork.VirtualNetwork{{
						Name: to.Ptr(name),
						Tags: map[string]*string{
							cloud.TagUID: to.Ptr("uid"),
						},
						Properties: &armnetwork.VirtualNetworkPropertiesFormat{
							Subnets: []*armnetwork.Subnet{
								{Properties: &armnetwork.SubnetPropertiesFormat{AddressPrefix: to.Ptr(subnetworkCIDR)}},
							},
						},
					}},
				},
			},
			wantNetworkCIDR: subnetworkCIDR,
		},
		"no virtual networks found": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			virtualNetworksAPI: &stubVirtualNetworksAPI{
				pager: &stubVirtualNetworksClientListPager{},
			},
			wantErr:         true,
			wantNetworkCIDR: subnetworkCIDR,
		},
		"malformed network struct": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			virtualNetworksAPI: &stubVirtualNetworksAPI{
				pager: &stubVirtualNetworksClientListPager{list: []armnetwork.VirtualNetwork{{}}},
			},
			wantErr:         true,
			wantNetworkCIDR: subnetworkCIDR,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Cloud{
				imds:       tc.imdsAPI,
				virtNetAPI: tc.virtualNetworksAPI,
			}
			subnetworkCIDR, err := metadata.getSubnetworkCIDR(t.Context())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNetworkCIDR, subnetworkCIDR)
		})
	}
}

func TestGetLoadBalancerEndpoint(t *testing.T) {
	someErr := errors.New("failed")
	goodLB := armnetwork.LoadBalancer{
		Name: to.Ptr("load-balancer"),
		Tags: map[string]*string{
			cloud.TagUID: to.Ptr("uid"),
		},
		Properties: &armnetwork.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
				{
					Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
					},
				},
			},
		},
	}
	goodIP := armnetwork.PublicIPAddressesClientGetResponse{
		PublicIPAddress: armnetwork.PublicIPAddress{
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				IPAddress: to.Ptr("192.0.2.1"),
			},
		},
	}

	testCases := map[string]struct {
		loadBalancerAPI      loadBalancerAPI
		publicIPAddressesAPI publicIPAddressesAPI
		imdsAPI              imdsAPI
		wantIP               string
		wantErr              bool
	}{
		"GetLoadBalancerEndpoint works": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: goodIP},
			wantIP:               "192.0.2.1",
		},
		"incorrect uid": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
						Name: to.Ptr("load-balancer"),
						Tags: map[string]*string{
							cloud.TagUID: to.Ptr("not-the-searched-uid"),
						},
						Properties: &armnetwork.LoadBalancerPropertiesFormat{
							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
								{
									Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
										PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
									},
								},
							},
						},
					}},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: goodIP},
			wantErr:              true,
		},
		"no load balancer": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: goodIP},
			wantErr:              true,
		},
		"no public IP address found": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getErr: someErr},
			wantErr:              true,
		},
		"found public IP has no address field": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: armnetwork.PublicIPAddressesClientGetResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{
					Properties: &armnetwork.PublicIPAddressPropertiesFormat{},
				},
			}},
			wantErr: true,
		},
		"imds resource group error": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupErr: someErr,
				uidVal:           "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: goodIP},
			wantErr:              true,
		},
		"imds uid error": {
			imdsAPI: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidErr:           someErr,
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			publicIPAddressesAPI: &stubPublicIPAddressesAPI{getResponse: goodIP},
			wantErr:              true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			metadata := Cloud{
				imds:            tc.imdsAPI,
				loadBalancerAPI: tc.loadBalancerAPI,
				pubIPAPI:        tc.publicIPAddressesAPI,
			}
			gotHost, gotPort, err := metadata.GetLoadBalancerEndpoint(t.Context())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantIP, gotHost)
			assert.Equal("6443", gotPort)
		})
	}
}

type stubIMDSAPI struct {
	providerIDErr     error
	providerIDVal     string
	subscriptionIDErr error
	subscriptionIDVal string
	resourceGroupErr  error
	resourceGroupVal  string
	uidErr            error
	uidVal            string
	nameErr           error
	nameVal           string
	initSecretHashVal string
	initSecretHashErr error
}

func (a *stubIMDSAPI) providerID(_ context.Context) (string, error) {
	return a.providerIDVal, a.providerIDErr
}

func (a *stubIMDSAPI) subscriptionID(_ context.Context) (string, error) {
	return a.subscriptionIDVal, a.subscriptionIDErr
}

func (a *stubIMDSAPI) resourceGroup(_ context.Context) (string, error) {
	return a.resourceGroupVal, a.resourceGroupErr
}

func (a *stubIMDSAPI) uid(_ context.Context) (string, error) {
	return a.uidVal, a.uidErr
}

func (a *stubIMDSAPI) name(_ context.Context) (string, error) {
	return a.nameVal, a.nameErr
}

func (a *stubIMDSAPI) initSecretHash(_ context.Context) (string, error) {
	return a.initSecretHashVal, a.initSecretHashErr
}

type stubVirtualMachineScaleSetVMPager struct {
	list     []armcompute.VirtualMachineScaleSetVM
	fetchErr error
	more     bool
}

func (p *stubVirtualMachineScaleSetVMPager) moreFunc() func(armcompute.VirtualMachineScaleSetVMsClientListResponse) bool {
	return func(armcompute.VirtualMachineScaleSetVMsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualMachineScaleSetVMPager) fetcherFunc() func(context.Context, *armcompute.VirtualMachineScaleSetVMsClientListResponse,
) (armcompute.VirtualMachineScaleSetVMsClientListResponse, error) {
	return func(context.Context, *armcompute.VirtualMachineScaleSetVMsClientListResponse) (armcompute.VirtualMachineScaleSetVMsClientListResponse, error) {
		page := make([]*armcompute.VirtualMachineScaleSetVM, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcompute.VirtualMachineScaleSetVMsClientListResponse{
			VirtualMachineScaleSetVMListResult: armcompute.VirtualMachineScaleSetVMListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubVirtualMachineScaleSetVMsAPI struct {
	getVM  armcompute.VirtualMachineScaleSetVM
	getErr error
	pager  *stubVirtualMachineScaleSetVMPager
}

func (a *stubVirtualMachineScaleSetVMsAPI) Get(context.Context, string, string, string, *armcompute.VirtualMachineScaleSetVMsClientGetOptions,
) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return armcompute.VirtualMachineScaleSetVMsClientGetResponse{
		VirtualMachineScaleSetVM: a.getVM,
	}, a.getErr
}

func (a *stubVirtualMachineScaleSetVMsAPI) NewListPager(string, string, *armcompute.VirtualMachineScaleSetVMsClientListOptions,
) *runtime.Pager[armcompute.VirtualMachineScaleSetVMsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachineScaleSetVMsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubNetworkInterfacesAPI struct {
	getInterface armnetwork.Interface
	getErr       error
}

func (a *stubNetworkInterfacesAPI) GetVirtualMachineScaleSetNetworkInterface(context.Context, string,
	string, string, string, *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error) {
	return armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse{
		Interface: a.getInterface,
	}, a.getErr
}

func (a *stubNetworkInterfacesAPI) Get(context.Context, string, string, *armnetwork.InterfacesClientGetOptions,
) (armnetwork.InterfacesClientGetResponse, error) {
	return armnetwork.InterfacesClientGetResponse{
		Interface: a.getInterface,
	}, a.getErr
}

type stubVirtualMachineScaleSetsClientListPager struct {
	list     []armcompute.VirtualMachineScaleSet
	fetchErr error
	more     bool
}

func (p *stubVirtualMachineScaleSetsClientListPager) moreFunc() func(armcompute.VirtualMachineScaleSetsClientListResponse) bool {
	return func(armcompute.VirtualMachineScaleSetsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualMachineScaleSetsClientListPager) fetcherFunc() func(context.Context, *armcompute.VirtualMachineScaleSetsClientListResponse,
) (armcompute.VirtualMachineScaleSetsClientListResponse, error) {
	return func(context.Context, *armcompute.VirtualMachineScaleSetsClientListResponse) (armcompute.VirtualMachineScaleSetsClientListResponse, error) {
		page := make([]*armcompute.VirtualMachineScaleSet, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcompute.VirtualMachineScaleSetsClientListResponse{
			VirtualMachineScaleSetListResult: armcompute.VirtualMachineScaleSetListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubScaleSetsAPI struct {
	pager *stubVirtualMachineScaleSetsClientListPager
}

func (a *stubScaleSetsAPI) NewListPager(string, *armcompute.VirtualMachineScaleSetsClientListOptions,
) *runtime.Pager[armcompute.VirtualMachineScaleSetsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachineScaleSetsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubSecurityGroupsClientListPager struct {
	list     []armnetwork.SecurityGroup
	fetchErr error
	more     bool
}

func (p *stubSecurityGroupsClientListPager) moreFunc() func(armnetwork.SecurityGroupsClientListResponse) bool {
	return func(armnetwork.SecurityGroupsClientListResponse) bool {
		return p.more
	}
}

func (p *stubSecurityGroupsClientListPager) fetcherFunc() func(context.Context, *armnetwork.SecurityGroupsClientListResponse,
) (armnetwork.SecurityGroupsClientListResponse, error) {
	return func(context.Context, *armnetwork.SecurityGroupsClientListResponse) (armnetwork.SecurityGroupsClientListResponse, error) {
		page := make([]*armnetwork.SecurityGroup, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armnetwork.SecurityGroupsClientListResponse{
			SecurityGroupListResult: armnetwork.SecurityGroupListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubSecurityGroupsAPI struct {
	pager *stubSecurityGroupsClientListPager
}

func (a *stubSecurityGroupsAPI) NewListPager(string, *armnetwork.SecurityGroupsClientListOptions,
) *runtime.Pager[armnetwork.SecurityGroupsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armnetwork.SecurityGroupsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubVirtualNetworksClientListPager struct {
	list     []armnetwork.VirtualNetwork
	fetchErr error
	more     bool
}

func (p *stubVirtualNetworksClientListPager) moreFunc() func(armnetwork.VirtualNetworksClientListResponse) bool {
	return func(armnetwork.VirtualNetworksClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualNetworksClientListPager) fetcherFunc() func(context.Context, *armnetwork.VirtualNetworksClientListResponse,
) (armnetwork.VirtualNetworksClientListResponse, error) {
	return func(context.Context, *armnetwork.VirtualNetworksClientListResponse) (armnetwork.VirtualNetworksClientListResponse, error) {
		page := make([]*armnetwork.VirtualNetwork, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armnetwork.VirtualNetworksClientListResponse{
			VirtualNetworkListResult: armnetwork.VirtualNetworkListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubVirtualNetworksAPI struct {
	pager *stubVirtualNetworksClientListPager
}

func (a *stubVirtualNetworksAPI) NewListPager(string, *armnetwork.VirtualNetworksClientListOptions,
) *runtime.Pager[armnetwork.VirtualNetworksClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armnetwork.VirtualNetworksClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubLoadBalancersClientListPager struct {
	list     []armnetwork.LoadBalancer
	fetchErr error
	more     bool
}

func (p *stubLoadBalancersClientListPager) moreFunc() func(armnetwork.LoadBalancersClientListResponse) bool {
	return func(armnetwork.LoadBalancersClientListResponse) bool {
		return p.more
	}
}

func (p *stubLoadBalancersClientListPager) fetcherFunc() func(context.Context, *armnetwork.LoadBalancersClientListResponse,
) (armnetwork.LoadBalancersClientListResponse, error) {
	return func(context.Context, *armnetwork.LoadBalancersClientListResponse) (armnetwork.LoadBalancersClientListResponse, error) {
		page := make([]*armnetwork.LoadBalancer, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armnetwork.LoadBalancersClientListResponse{
			LoadBalancerListResult: armnetwork.LoadBalancerListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubLoadBalancersAPI struct {
	pager *stubLoadBalancersClientListPager
}

func (a *stubLoadBalancersAPI) NewListPager(_ string, _ *armnetwork.LoadBalancersClientListOptions,
) *runtime.Pager[armnetwork.LoadBalancersClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armnetwork.LoadBalancersClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubPublicIPAddressesAPI struct {
	getResponse                                      armnetwork.PublicIPAddressesClientGetResponse
	getVirtualMachineScaleSetPublicIPAddressResponse armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse
	getErr                                           error
}

func (a *stubPublicIPAddressesAPI) Get(context.Context, string, string, *armnetwork.PublicIPAddressesClientGetOptions,
) (armnetwork.PublicIPAddressesClientGetResponse, error) {
	return a.getResponse, a.getErr
}

func (a *stubPublicIPAddressesAPI) GetVirtualMachineScaleSetPublicIPAddress(context.Context, string, string,
	string, string, string, string, *armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressOptions,
) (armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse, error) {
	return a.getVirtualMachineScaleSetPublicIPAddressResponse, a.getErr
}
