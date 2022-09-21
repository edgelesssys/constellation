/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"crypto/rand"
	"math/big"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azure"
)

// ScaleSet defines a Azure scale set.
type ScaleSet struct {
	Name                           string
	NamePrefix                     string
	UID                            string
	Subscription                   string
	ResourceGroup                  string
	Location                       string
	InstanceType                   string
	StateDiskSizeGB                int32
	StateDiskType                  string
	Count                          int64
	Username                       string
	SubnetID                       string
	NetworkSecurityGroup           string
	Password                       string
	Image                          string
	UserAssignedIdentity           string
	LoadBalancerName               string
	LoadBalancerBackendAddressPool string
	ConfidentialVM                 bool
}

// Azure returns the Azure representation of ScaleSet.
func (s ScaleSet) Azure() armcomputev2.VirtualMachineScaleSet {
	securityType := armcomputev2.SecurityTypesTrustedLaunch
	var diskSecurityProfile *armcomputev2.VMDiskSecurityProfile
	if s.ConfidentialVM {
		securityType = armcomputev2.SecurityTypesConfidentialVM
		diskSecurityProfile = &armcomputev2.VMDiskSecurityProfile{
			SecurityEncryptionType: to.Ptr(armcomputev2.SecurityEncryptionTypesVMGuestStateOnly),
		}
	}

	return armcomputev2.VirtualMachineScaleSet{
		Name:     to.Ptr(s.Name),
		Location: to.Ptr(s.Location),
		SKU: &armcomputev2.SKU{
			Name:     to.Ptr(s.InstanceType),
			Capacity: to.Ptr(s.Count),
		},
		Properties: &armcomputev2.VirtualMachineScaleSetProperties{
			Overprovision: to.Ptr(false),
			UpgradePolicy: &armcomputev2.UpgradePolicy{
				Mode: to.Ptr(armcomputev2.UpgradeModeManual),
				AutomaticOSUpgradePolicy: &armcomputev2.AutomaticOSUpgradePolicy{
					EnableAutomaticOSUpgrade: to.Ptr(false),
					DisableAutomaticRollback: to.Ptr(false),
				},
			},
			VirtualMachineProfile: &armcomputev2.VirtualMachineScaleSetVMProfile{
				OSProfile: &armcomputev2.VirtualMachineScaleSetOSProfile{
					ComputerNamePrefix: to.Ptr(s.NamePrefix),
					AdminUsername:      to.Ptr(s.Username),
					AdminPassword:      to.Ptr(s.Password),
					LinuxConfiguration: &armcomputev2.LinuxConfiguration{},
				},
				StorageProfile: &armcomputev2.VirtualMachineScaleSetStorageProfile{
					ImageReference: azure.ImageReferenceFromImage(s.Image),
					DataDisks: []*armcomputev2.VirtualMachineScaleSetDataDisk{
						{
							CreateOption: to.Ptr(armcomputev2.DiskCreateOptionTypesEmpty),
							DiskSizeGB:   to.Ptr(s.StateDiskSizeGB),
							Lun:          to.Ptr[int32](0),
							ManagedDisk: &armcomputev2.VirtualMachineScaleSetManagedDiskParameters{
								StorageAccountType: (*armcomputev2.StorageAccountTypes)(to.Ptr(s.StateDiskType)),
							},
						},
					},
					OSDisk: &armcomputev2.VirtualMachineScaleSetOSDisk{
						ManagedDisk: &armcomputev2.VirtualMachineScaleSetManagedDiskParameters{
							SecurityProfile: diskSecurityProfile,
						},
						CreateOption: to.Ptr(armcomputev2.DiskCreateOptionTypesFromImage),
					},
				},
				NetworkProfile: &armcomputev2.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: []*armcomputev2.VirtualMachineScaleSetNetworkConfiguration{
						{
							Name: to.Ptr(s.Name),
							Properties: &armcomputev2.VirtualMachineScaleSetNetworkConfigurationProperties{
								Primary:            to.Ptr(true),
								EnableIPForwarding: to.Ptr(true),
								IPConfigurations: []*armcomputev2.VirtualMachineScaleSetIPConfiguration{
									{
										Name: to.Ptr(s.Name),
										Properties: &armcomputev2.VirtualMachineScaleSetIPConfigurationProperties{
											Primary: to.Ptr(true),
											Subnet: &armcomputev2.APIEntityReference{
												ID: to.Ptr(s.SubnetID),
											},
											LoadBalancerBackendAddressPools: []*armcomputev2.SubResource{
												{
													ID: to.Ptr("/subscriptions/" + s.Subscription +
														"/resourcegroups/" + s.ResourceGroup +
														"/providers/Microsoft.Network/loadBalancers/" + s.LoadBalancerName +
														"/backendAddressPools/" + s.LoadBalancerBackendAddressPool),
												},
												{
													ID: to.Ptr("/subscriptions/" + s.Subscription +
														"/resourcegroups/" + s.ResourceGroup +
														"/providers/Microsoft.Network/loadBalancers/" + s.LoadBalancerName +
														"/backendAddressPools/all"),
												},
											},
										},
									},
								},
								NetworkSecurityGroup: &armcomputev2.SubResource{
									ID: to.Ptr(s.NetworkSecurityGroup),
								},
							},
						},
					},
				},
				SecurityProfile: &armcomputev2.SecurityProfile{
					SecurityType: to.Ptr(securityType),
					UefiSettings: &armcomputev2.UefiSettings{VTpmEnabled: to.Ptr(true), SecureBootEnabled: to.Ptr(true)},
				},
				DiagnosticsProfile: &armcomputev2.DiagnosticsProfile{
					BootDiagnostics: &armcomputev2.BootDiagnostics{
						Enabled: to.Ptr(true),
					},
				},
			},
		},
		Identity: &armcomputev2.VirtualMachineScaleSetIdentity{
			Type: to.Ptr(armcomputev2.ResourceIdentityTypeUserAssigned),
			UserAssignedIdentities: map[string]*armcomputev2.UserAssignedIdentitiesValue{
				s.UserAssignedIdentity: {},
			},
		},
		Tags: map[string]*string{"uid": to.Ptr(s.UID)},
	}
}

// GeneratePassword is a helper function to generate a random password
// for Azure's scale set.
func GeneratePassword() (string, error) {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	pwLen := 16
	pw := make([]byte, 0, pwLen)
	for i := 0; i < pwLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		pw = append(pw, letters[n.Int64()])
	}
	// bypass password rules
	pw = append(pw, []byte("Aa1!")...)
	return string(pw), nil
}
