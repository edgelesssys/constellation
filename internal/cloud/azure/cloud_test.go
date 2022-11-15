/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGetCCMConfig(t *testing.T) {
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
	goodSecurityGroup := armnetwork.SecurityGroup{
		Tags: map[string]*string{
			cloud.TagUID: to.Ptr("uid"),
		},
		Name: to.Ptr("security-group"),
	}

	testCases := map[string]struct {
		imdsAPI                imdsAPI
		loadBalancerAPI        loadBalancerAPI
		secGroupAPI            securityGroupsAPI
		providerID             string
		cloudServiceAccountURI string
		wantErr                bool
		wantConfig             cloudConfig
	}{
		"success": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantConfig: cloudConfig{
				Cloud:               "AzurePublicCloud",
				TenantID:            "tenant-id",
				SubscriptionID:      "subscription-id",
				ResourceGroup:       "resource-group",
				LoadBalancerSku:     "standard",
				SecurityGroupName:   "security-group",
				LoadBalancerName:    "load-balancer",
				UseInstanceMetadata: true,
				VMType:              "vmss",
				Location:            "westeurope",
				AADClientID:         "client-id",
				AADClientSecret:     "client-secret",
			},
		},
		"missing UID tag": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{
						{
							Name: to.Ptr("load-balancer"),
							Properties: &armnetwork.LoadBalancerPropertiesFormat{
								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
									{
										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
											PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
										},
									},
								},
							},
						},
					},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
		"only correct UID is chosen": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{
						{
							Name: to.Ptr("load-balancer"),
							Tags: map[string]*string{
								cloud.TagUID: to.Ptr("different-uid"),
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
						},
						goodLB,
					},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantConfig: cloudConfig{
				Cloud:               "AzurePublicCloud",
				TenantID:            "tenant-id",
				SubscriptionID:      "subscription-id",
				ResourceGroup:       "resource-group",
				LoadBalancerSku:     "standard",
				SecurityGroupName:   "security-group",
				LoadBalancerName:    "load-balancer",
				UseInstanceMetadata: true,
				VMType:              "vmss",
				Location:            "westeurope",
				AADClientID:         "client-id",
				AADClientSecret:     "client-secret",
			},
		},
		"load balancer list error": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					fetchErr: someErr,
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
		"missing load balancer name": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{{
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
					}},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
		"security group list error": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					fetchErr: someErr,
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
		"invalid provider ID": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "invalid:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
		"invalid cloud service account URI": {
			imdsAPI: &stubIMDSAPI{
				uidVal: "uid",
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "invalid://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
		"imds error": {
			imdsAPI: &stubIMDSAPI{
				uidErr: someErr,
			},
			loadBalancerAPI: &stubLoadBalancersAPI{
				pager: &stubLoadBalancersClientListPager{
					list: []armnetwork.LoadBalancer{goodLB},
				},
			},
			secGroupAPI: &stubSecurityGroupsAPI{
				pager: &stubSecurityGroupsClientListPager{
					list: []armnetwork.SecurityGroup{goodSecurityGroup},
				},
			},
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cloud := &Cloud{
				imds:            tc.imdsAPI,
				loadBalancerAPI: tc.loadBalancerAPI,
				secGroupAPI:     tc.secGroupAPI,
			}
			config, err := cloud.GetCCMConfig(context.Background(), tc.providerID, tc.cloudServiceAccountURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			wantConfig, err := json.Marshal(tc.wantConfig)
			require.NoError(err)
			assert.JSONEq(string(wantConfig), string(config))
		})
	}
}

func TestGetInstance(t *testing.T) {
	someErr := errors.New("failed")
	sampleProviderID := "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"
	successVMAPI := &stubVirtualMachineScaleSetVMsAPI{
		getVM: armcomputev2.VirtualMachineScaleSetVM{
			Name: to.Ptr("scale-set-name-instance-id"),
			ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
			Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
				OSProfile: &armcomputev2.OSProfile{
					ComputerName: to.Ptr("scale-set-name-instance-id"),
				},
				NetworkProfile: &armcomputev2.NetworkProfile{
					NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
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
		providerID           string
		wantInstance         metadata.InstanceMetadata
		wantErr              bool
	}{
		"success worker": {
			scaleSetsVMAPI:       successVMAPI,
			networkInterfacesAPI: successNetworkAPI,
			providerID:           sampleProviderID,
			wantInstance: metadata.InstanceMetadata{
				Name:       "scale-set-name-instance-id",
				ProviderID: sampleProviderID,
				Role:       role.Worker,
				VPCIP:      "192.0.2.1",
			},
		},
		"success control-plane": {
			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
				getVM: armcomputev2.VirtualMachineScaleSetVM{
					Name: to.Ptr("scale-set-name-instance-id"),
					ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"),
					Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
						OSProfile: &armcomputev2.OSProfile{
							ComputerName: to.Ptr("scale-set-name-instance-id"),
						},
						NetworkProfile: &armcomputev2.NetworkProfile{
							NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
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
			networkInterfacesAPI: successNetworkAPI,
			providerID:           "invalid",
			wantErr:              true,
		},
		"vm API error": {
			scaleSetsVMAPI:       &stubVirtualMachineScaleSetVMsAPI{getErr: someErr},
			networkInterfacesAPI: successNetworkAPI,
			providerID:           sampleProviderID,
			wantErr:              true,
		},
		"network API error": {
			scaleSetsVMAPI:       successVMAPI,
			networkInterfacesAPI: &stubNetworkInterfacesAPI{getErr: someErr},
			providerID:           sampleProviderID,
			wantErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			metadata := Cloud{
				scaleSetsVMAPI: tc.scaleSetsVMAPI,
				netIfacAPI:     tc.networkInterfacesAPI,
			}
			instance, err := metadata.getInstance(context.Background(), tc.providerID)
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
			uid, err := cloud.UID(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.imdsAPI.uidVal, uid)
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
			list: []armcomputev2.VirtualMachineScaleSet{{
				Name: to.Ptr("scale-set"),
				Tags: map[string]*string{
					cloud.TagUID:  to.Ptr("uid"),
					cloud.TagRole: to.Ptr("worker"),
				},
			}},
		},
	}
	scaleSetVM := armcomputev2.VirtualMachineScaleSetVM{
		Name:       to.Ptr("scale-set_0"),
		InstanceID: to.Ptr("instance-id"),
		ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0"),
		Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
			NetworkProfile: &armcomputev2.NetworkProfile{
				NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
					{
						ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0/networkInterfaces/interface-name"),
					},
				},
			},
			OSProfile: &armcomputev2.OSProfile{
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
					list: []armcomputev2.VirtualMachineScaleSetVM{scaleSetVM},
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
					list: []armcomputev2.VirtualMachineScaleSetVM{
						scaleSetVM,
						{
							Name:       to.Ptr("control-set_0"),
							InstanceID: to.Ptr("0"),
							ID:         to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/control-set/virtualMachines/0"),
							Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
								NetworkProfile: &armcomputev2.NetworkProfile{
									NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
										{
											ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/control-set/virtualMachines/0/networkInterfaces/interface-name"),
										},
									},
								},
								OSProfile: &armcomputev2.OSProfile{
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
					list: []armcomputev2.VirtualMachineScaleSetVM{scaleSetVM},
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
					list: []armcomputev2.VirtualMachineScaleSetVM{scaleSetVM},
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
			instances, err := azureMetadata.List(context.Background())

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
			name, err := metadata.getNetworkSecurityGroupName(context.Background(), "resource-group", "uid")
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
			subnetworkCIDR, err := metadata.getSubnetworkCIDR(context.Background())
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
			loadbalancerName, err := metadata.GetLoadBalancerEndpoint(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantIP, loadbalancerName)
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
}

func (a *stubIMDSAPI) providerID(ctx context.Context) (string, error) {
	return a.providerIDVal, a.providerIDErr
}

func (a *stubIMDSAPI) subscriptionID(ctx context.Context) (string, error) {
	return a.subscriptionIDVal, a.subscriptionIDErr
}

func (a *stubIMDSAPI) resourceGroup(ctx context.Context) (string, error) {
	return a.resourceGroupVal, a.resourceGroupErr
}

func (a *stubIMDSAPI) uid(ctx context.Context) (string, error) {
	return a.uidVal, a.uidErr
}

func (a *stubIMDSAPI) name(ctx context.Context) (string, error) {
	return a.nameVal, a.nameErr
}

type stubVirtualMachineScaleSetVMPager struct {
	list     []armcomputev2.VirtualMachineScaleSetVM
	fetchErr error
	more     bool
}

func (p *stubVirtualMachineScaleSetVMPager) moreFunc() func(armcomputev2.VirtualMachineScaleSetVMsClientListResponse) bool {
	return func(armcomputev2.VirtualMachineScaleSetVMsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualMachineScaleSetVMPager) fetcherFunc() func(context.Context, *armcomputev2.VirtualMachineScaleSetVMsClientListResponse,
) (armcomputev2.VirtualMachineScaleSetVMsClientListResponse, error) {
	return func(context.Context, *armcomputev2.VirtualMachineScaleSetVMsClientListResponse) (armcomputev2.VirtualMachineScaleSetVMsClientListResponse, error) {
		page := make([]*armcomputev2.VirtualMachineScaleSetVM, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcomputev2.VirtualMachineScaleSetVMsClientListResponse{
			VirtualMachineScaleSetVMListResult: armcomputev2.VirtualMachineScaleSetVMListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubVirtualMachineScaleSetVMsAPI struct {
	getVM  armcomputev2.VirtualMachineScaleSetVM
	getErr error
	pager  *stubVirtualMachineScaleSetVMPager
}

func (a *stubVirtualMachineScaleSetVMsAPI) Get(context.Context, string, string, string, *armcomputev2.VirtualMachineScaleSetVMsClientGetOptions,
) (armcomputev2.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return armcomputev2.VirtualMachineScaleSetVMsClientGetResponse{
		VirtualMachineScaleSetVM: a.getVM,
	}, a.getErr
}

func (a *stubVirtualMachineScaleSetVMsAPI) NewListPager(string, string, *armcomputev2.VirtualMachineScaleSetVMsClientListOptions,
) *runtime.Pager[armcomputev2.VirtualMachineScaleSetVMsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcomputev2.VirtualMachineScaleSetVMsClientListResponse]{
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
	list     []armcomputev2.VirtualMachineScaleSet
	fetchErr error
	more     bool
}

func (p *stubVirtualMachineScaleSetsClientListPager) moreFunc() func(armcomputev2.VirtualMachineScaleSetsClientListResponse) bool {
	return func(armcomputev2.VirtualMachineScaleSetsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualMachineScaleSetsClientListPager) fetcherFunc() func(context.Context, *armcomputev2.VirtualMachineScaleSetsClientListResponse,
) (armcomputev2.VirtualMachineScaleSetsClientListResponse, error) {
	return func(context.Context, *armcomputev2.VirtualMachineScaleSetsClientListResponse) (armcomputev2.VirtualMachineScaleSetsClientListResponse, error) {
		page := make([]*armcomputev2.VirtualMachineScaleSet, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcomputev2.VirtualMachineScaleSetsClientListResponse{
			VirtualMachineScaleSetListResult: armcomputev2.VirtualMachineScaleSetListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubScaleSetsAPI struct {
	pager *stubVirtualMachineScaleSetsClientListPager
}

func (a *stubScaleSetsAPI) NewListPager(string, *armcomputev2.VirtualMachineScaleSetsClientListOptions,
) *runtime.Pager[armcomputev2.VirtualMachineScaleSetsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcomputev2.VirtualMachineScaleSetsClientListResponse]{
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

func (a *stubLoadBalancersAPI) NewListPager(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions,
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
