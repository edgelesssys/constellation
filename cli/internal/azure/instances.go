package azure

// copy of ec2/instances.go

// TODO(katexochen): refactor into mulitcloud package.

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
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
func (i VMInstance) Azure() armcomputev2.VirtualMachine {
	return armcomputev2.VirtualMachine{
		Name:     to.Ptr(i.Name),
		Location: to.Ptr(i.Location),
		Properties: &armcomputev2.VirtualMachineProperties{
			HardwareProfile: &armcomputev2.HardwareProfile{
				VMSize: (*armcomputev2.VirtualMachineSizeTypes)(to.Ptr(i.InstanceType)),
			},
			OSProfile: &armcomputev2.OSProfile{
				ComputerName:  to.Ptr(i.Name),
				AdminPassword: to.Ptr(i.Password),
				AdminUsername: to.Ptr(i.Username),
			},
			SecurityProfile: &armcomputev2.SecurityProfile{
				UefiSettings: &armcomputev2.UefiSettings{
					SecureBootEnabled: to.Ptr(true),
					VTpmEnabled:       to.Ptr(true),
				},
				SecurityType: to.Ptr(armcomputev2.SecurityTypesConfidentialVM),
			},
			NetworkProfile: &armcomputev2.NetworkProfile{
				NetworkInterfaces: []*armcomputev2.NetworkInterfaceReference{
					{
						ID: to.Ptr(i.NIC),
					},
				},
			},
			StorageProfile: &armcomputev2.StorageProfile{
				OSDisk: &armcomputev2.OSDisk{
					CreateOption: to.Ptr(armcomputev2.DiskCreateOptionTypesFromImage),
					ManagedDisk: &armcomputev2.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcomputev2.StorageAccountTypesPremiumLRS),
						SecurityProfile: &armcomputev2.VMDiskSecurityProfile{
							SecurityEncryptionType: to.Ptr(armcomputev2.SecurityEncryptionTypesVMGuestStateOnly),
						},
					},
				},
				ImageReference: &armcomputev2.ImageReference{
					Publisher: to.Ptr("0001-com-ubuntu-confidential-vm-focal"),
					Offer:     to.Ptr("canonical"),
					SKU:       to.Ptr("20_04-lts-gen2"),
					Version:   to.Ptr("latest"),
				},
			},
			DiagnosticsProfile: &armcomputev2.DiagnosticsProfile{
				BootDiagnostics: &armcomputev2.BootDiagnostics{
					Enabled: to.Ptr(true),
				},
			},
		},
		Identity: &armcomputev2.VirtualMachineIdentity{
			Type: to.Ptr(armcomputev2.ResourceIdentityTypeSystemAssigned),
		},
	}
}
