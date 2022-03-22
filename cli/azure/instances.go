package azure

// copy of ec2/instances.go

// TODO(katexochen): refactor into mulitcloud package.

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// Instance is a azure instance.
type Instance struct {
	PublicIP  string
	PrivateIP string
}

// Instances is a map of azure Instances. The ID of an instance is used as key.
type Instances map[string]Instance

// IDs returns the IDs of all instances of the Constellation.
func (i Instances) IDs() []string {
	var ids []string
	for id := range i {
		ids = append(ids, id)
	}
	return ids
}

// PublicIPs returns the public IPs of all the instances of the Constellation.
func (i Instances) PublicIPs() []string {
	var ips []string
	for _, instance := range i {
		ips = append(ips, instance.PublicIP)
	}
	return ips
}

// PrivateIPs returns the private IPs of all the instances of the Constellation.
func (i Instances) PrivateIPs() []string {
	var ips []string
	for _, instance := range i {
		ips = append(ips, instance.PrivateIP)
	}
	return ips
}

// GetOne return anyone instance out of the instances and its ID.
func (i Instances) GetOne() (string, Instance, error) {
	for id, instance := range i {
		return id, instance, nil
	}
	return "", Instance{}, errors.New("map is empty")
}

// GetOthers returns all instances but the one with the handed ID.
func (i Instances) GetOthers(id string) Instances {
	others := make(Instances)
	for key, instance := range i {
		if key != id {
			others[key] = instance
		}
	}
	return others
}

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
