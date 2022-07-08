package azure

// copy of ec2/instances.go

// TODO(katexochen): refactor into mulitcloud package.

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// VMInstance describes a single instance.
// TODO: deprecate as soon as scale sets are available.
type VMInstance struct {
	Name         string
	Location     string
	InstanceType string
	Username     string
	Password     string
	NIC          string
	Image        string
}

// Azure makes a new virtual machine template with default values.
// TODO: deprecate as soon as scale sets are available.
func (i VMInstance) Azure() armcompute.VirtualMachine {
	return armcompute.VirtualMachine{
		Name:     to.StringPtr(i.Name),
		Location: to.StringPtr(i.Location),
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: (*armcompute.VirtualMachineSizeTypes)(to.StringPtr(i.InstanceType)),
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  to.StringPtr(i.Name),
				AdminPassword: to.StringPtr(i.Password),
				AdminUsername: to.StringPtr(i.Username),
			},
			SecurityProfile: &armcompute.SecurityProfile{
				UefiSettings: &armcompute.UefiSettings{
					SecureBootEnabled: to.BoolPtr(true),
					VTpmEnabled:       to.BoolPtr(true),
				},
				SecurityType: armcompute.SecurityTypesConfidentialVM.ToPtr(),
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.StringPtr(i.NIC),
					},
				},
			},
			StorageProfile: &armcompute.StorageProfile{
				OSDisk: &armcompute.OSDisk{
					CreateOption: armcompute.DiskCreateOptionTypesFromImage.ToPtr(),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: armcompute.StorageAccountTypesPremiumLRS.ToPtr(),
						SecurityProfile: &armcompute.VMDiskSecurityProfile{
							SecurityEncryptionType: armcompute.SecurityEncryptionTypesVMGuestStateOnly.ToPtr(),
						},
					},
				},
				ImageReference: &armcompute.ImageReference{
					Publisher: to.StringPtr("0001-com-ubuntu-confidential-vm-focal"),
					Offer:     to.StringPtr("canonical"),
					SKU:       to.StringPtr("20_04-lts-gen2"),
					Version:   to.StringPtr("latest"),
				},
			},
			DiagnosticsProfile: &armcompute.DiagnosticsProfile{
				BootDiagnostics: &armcompute.BootDiagnostics{
					Enabled: to.BoolPtr(true),
				},
			},
		},
		Identity: &armcompute.VirtualMachineIdentity{
			Type: armcompute.ResourceIdentityTypeSystemAssigned.ToPtr(),
		},
	}
}
