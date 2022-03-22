package azure

import "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"

// InstanceTypes are valid Azure instance types.
// Normally, this would be string(armcompute.VirtualMachineSizeTypesStandardD4SV3),
// but currently needed instances are not in SDK.
var InstanceTypes = []string{
	string(armcompute.VirtualMachineSizeTypesStandardD4SV3),
	"Standard_DC2as_v5",
	"Standard_DC4as_v5",
	"Standard_DC8as_v5",
}
