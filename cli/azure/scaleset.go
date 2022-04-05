package azure

import (
	"crypto/rand"
	"math/big"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// ScaleSet defines a Azure scale set.
type ScaleSet struct {
	Name                 string
	NamePrefix           string
	Location             string
	InstanceType         string
	StateDiskSizeGB      int32
	Count                int64
	Username             string
	SubnetID             string
	NetworkSecurityGroup string
	Password             string
	Image                string
	UserAssignedIdentity string
}

// Azure returns the Azure representation of ScaleSet.
func (s ScaleSet) Azure() armcompute.VirtualMachineScaleSet {
	return armcompute.VirtualMachineScaleSet{
		Name:     to.StringPtr(s.Name),
		Location: to.StringPtr(s.Location),
		SKU: &armcompute.SKU{
			Name:     to.StringPtr(s.InstanceType),
			Capacity: to.Int64Ptr(s.Count),
		},
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			Overprovision: to.BoolPtr(false),
			UpgradePolicy: &armcompute.UpgradePolicy{
				Mode: armcompute.UpgradeModeManual.ToPtr(),
				AutomaticOSUpgradePolicy: &armcompute.AutomaticOSUpgradePolicy{
					EnableAutomaticOSUpgrade: to.BoolPtr(false),
					DisableAutomaticRollback: to.BoolPtr(false),
				},
			},
			VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
				OSProfile: &armcompute.VirtualMachineScaleSetOSProfile{
					ComputerNamePrefix: to.StringPtr(s.NamePrefix),
					AdminUsername:      to.StringPtr(s.Username),
					AdminPassword:      to.StringPtr(s.Password),
					LinuxConfiguration: &armcompute.LinuxConfiguration{},
				},
				StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
					ImageReference: &armcompute.ImageReference{
						ID: to.StringPtr(s.Image),
					},
					DataDisks: []*armcompute.VirtualMachineScaleSetDataDisk{
						{
							CreateOption: armcompute.DiskCreateOptionTypesEmpty.ToPtr(),
							DiskSizeGB:   to.Int32Ptr(s.StateDiskSizeGB),
							Lun:          to.Int32Ptr(0),
						},
					},
				},
				NetworkProfile: &armcompute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: []*armcompute.VirtualMachineScaleSetNetworkConfiguration{
						{
							Name: to.StringPtr(s.Name),
							Properties: &armcompute.VirtualMachineScaleSetNetworkConfigurationProperties{
								Primary:            to.BoolPtr(true),
								EnableIPForwarding: to.BoolPtr(true),
								IPConfigurations: []*armcompute.VirtualMachineScaleSetIPConfiguration{
									{
										Name: to.StringPtr(s.Name),
										Properties: &armcompute.VirtualMachineScaleSetIPConfigurationProperties{
											Subnet: &armcompute.APIEntityReference{
												ID: to.StringPtr(s.SubnetID),
											},
											PublicIPAddressConfiguration: &armcompute.VirtualMachineScaleSetPublicIPAddressConfiguration{
												Name: to.StringPtr(s.Name),
												Properties: &armcompute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
													IdleTimeoutInMinutes: to.Int32Ptr(15), // default per https://docs.microsoft.com/en-us/azure/virtual-machine-scale-sets/virtual-machine-scale-sets-networking#creating-a-scale-set-with-public-ip-per-virtual-machine
												},
											},
										},
									},
								},
								NetworkSecurityGroup: &armcompute.SubResource{
									ID: to.StringPtr(s.NetworkSecurityGroup),
								},
							},
						},
					},
				},
				SecurityProfile: &armcompute.SecurityProfile{
					SecurityType: armcompute.SecurityTypesTrustedLaunch.ToPtr(),
					UefiSettings: &armcompute.UefiSettings{VTpmEnabled: to.BoolPtr(true)},
				},
				DiagnosticsProfile: &armcompute.DiagnosticsProfile{
					BootDiagnostics: &armcompute.BootDiagnostics{
						Enabled: to.BoolPtr(true),
					},
				},
			},
		},
		Identity: &armcompute.VirtualMachineScaleSetIdentity{
			Type: armcompute.ResourceIdentityTypeUserAssigned.ToPtr(),
			UserAssignedIdentities: map[string]*armcompute.VirtualMachineScaleSetIdentityUserAssignedIdentitiesValue{
				s.UserAssignedIdentity: {},
			},
		},
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
