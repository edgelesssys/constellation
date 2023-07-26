/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

// TODO(elchead): migrate to new location
// func TestGetCCMConfig(t *testing.T) {
//	someErr := errors.New("failed")
//	goodLB := armnetwork.LoadBalancer{
//		Name: to.Ptr("load-balancer"),
//		Tags: map[string]*string{
//			cloud.TagUID: to.Ptr("uid"),
//		},
//		Properties: &armnetwork.LoadBalancerPropertiesFormat{
//			FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
//				{
//					Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
//						PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
//					},
//				},
//			},
//		},
//	}
//	goodSecurityGroup := armnetwork.SecurityGroup{
//		Tags: map[string]*string{
//			cloud.TagUID: to.Ptr("uid"),
//		},
//		Name: to.Ptr("security-group"),
//	}

//	uamiClientID := "uami-client-id"

//	testCases := map[string]struct {
//		imdsAPI                imdsAPI
//		loadBalancerAPI        loadBalancerAPI
//		secGroupAPI            securityGroupsAPI
//		scaleSetsVMAPI         virtualMachineScaleSetVMsAPI
//		providerID             string
//		cloudServiceAccountURI string
//		wantErr                bool
//		wantConfig             cloudConfig
//	}{
//		"success": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{goodLB},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			scaleSetsVMAPI: &stubVirtualMachineScaleSetVMsAPI{
//				getVM: armcompute.VirtualMachineScaleSetVM{
//					Identity: &armcompute.VirtualMachineIdentity{
//						UserAssignedIdentities: map[string]*armcompute.UserAssignedIdentitiesValue{
//							"subscriptions/9b352db0-82af-408c-a02c-36fbffbf7015/resourceGroups/resourceGroupName/providers/Microsoft.ManagedIdentity/userAssignedIdentities/UAMIName": {ClientID: &uamiClientID},
//						},
//					},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope&preferred_auth_method=userassignedidentity&uami_resource_id=subscriptions%2F9b352db0-82af-408c-a02c-36fbffbf7015%2FresourceGroups%2FresourceGroupName%2Fproviders%2FMicrosoft.ManagedIdentity%2FuserAssignedIdentities%2FUAMIName",
//			wantConfig: cloudConfig{
//				Cloud:                       "AzurePublicCloud",
//				TenantID:                    "tenant-id",
//				SubscriptionID:              "subscription-id",
//				ResourceGroup:               "resource-group",
//				LoadBalancerSku:             "standard",
//				SecurityGroupName:           "security-group",
//				LoadBalancerName:            "load-balancer",
//				UseInstanceMetadata:         true,
//				UseManagedIdentityExtension: true,
//				UserAssignedIdentityID:      uamiClientID,
//				VMType:                      "vmss",
//				Location:                    "westeurope",
//			},
//		},
//		"no app registration": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{goodLB},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&location=westeurope",
//			wantConfig: cloudConfig{
//				Cloud:               "AzurePublicCloud",
//				TenantID:            "tenant-id",
//				SubscriptionID:      "subscription-id",
//				ResourceGroup:       "resource-group",
//				LoadBalancerSku:     "standard",
//				SecurityGroupName:   "security-group",
//				LoadBalancerName:    "load-balancer",
//				UseInstanceMetadata: true,
//				VMType:              "vmss",
//				Location:            "westeurope",
//			},
//		},
//		"missing UID tag": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{
//						{
//							Name: to.Ptr("load-balancer"),
//							Properties: &armnetwork.LoadBalancerPropertiesFormat{
//								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
//									{
//										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
//											PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//		"only correct UID is chosen": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{
//						{
//							Name: to.Ptr("load-balancer"),
//							Tags: map[string]*string{
//								cloud.TagUID: to.Ptr("different-uid"),
//							},
//							Properties: &armnetwork.LoadBalancerPropertiesFormat{
//								FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
//									{
//										Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
//											PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
//										},
//									},
//								},
//							},
//						},
//						goodLB,
//					},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantConfig: cloudConfig{
//				Cloud:               "AzurePublicCloud",
//				TenantID:            "tenant-id",
//				SubscriptionID:      "subscription-id",
//				ResourceGroup:       "resource-group",
//				LoadBalancerSku:     "standard",
//				SecurityGroupName:   "security-group",
//				LoadBalancerName:    "load-balancer",
//				UseInstanceMetadata: true,
//				VMType:              "vmss",
//				Location:            "westeurope",
//			},
//		},
//		"load balancer list error": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					fetchErr: someErr,
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//		"missing load balancer name": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{{
//						Tags: map[string]*string{
//							cloud.TagUID: to.Ptr("uid"),
//						},
//						Properties: &armnetwork.LoadBalancerPropertiesFormat{
//							FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
//								{
//									Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
//										PublicIPAddress: &armnetwork.PublicIPAddress{ID: to.Ptr("/subscriptions/subscription/resourceGroups/resourceGroup/providers/Microsoft.Network/publicIPAddresses/pubIPName")},
//									},
//								},
//							},
//						},
//					}},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//		"security group list error": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{goodLB},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					fetchErr: someErr,
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//		"invalid provider ID": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{goodLB},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "invalid:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//		"invalid cloud service account URI": {
//			imdsAPI: &stubIMDSAPI{
//				uidVal: "uid",
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{goodLB},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "invalid://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//		"imds error": {
//			imdsAPI: &stubIMDSAPI{
//				uidErr: someErr,
//			},
//			loadBalancerAPI: &stubLoadBalancersAPI{
//				pager: &stubLoadBalancersClientListPager{
//					list: []armnetwork.LoadBalancer{goodLB},
//				},
//			},
//			secGroupAPI: &stubSecurityGroupsAPI{
//				pager: &stubSecurityGroupsClientListPager{
//					list: []armnetwork.SecurityGroup{goodSecurityGroup},
//				},
//			},
//			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set/virtualMachines/0",
//			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=westeurope",
//			wantErr:                true,
//		},
//	}

//	for name, tc := range testCases {
//		t.Run(name, func(t *testing.T) {
//			assert := assert.New(t)
//			require := require.New(t)

//			cloud := &Cloud{
//				imds:            tc.imdsAPI,
//				loadBalancerAPI: tc.loadBalancerAPI,
//				secGroupAPI:     tc.secGroupAPI,
//				scaleSetsVMAPI:  tc.scaleSetsVMAPI,
//			}
//			config, err := cloud.GetCCMConfig(context.Background(), tc.providerID, tc.cloudServiceAccountURI)
//			if tc.wantErr {
//				assert.Error(err)
//				return
//			}
//			assert.NoError(err)

//			wantConfig, err := json.Marshal(tc.wantConfig)
//			require.NoError(err)
//			assert.JSONEq(string(wantConfig), string(config))
//		})
//	}
//}
