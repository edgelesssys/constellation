package azure

import "github.com/edgelesssys/constellation/coordinator/cloudprovider"

// CloudNodeManager holds the Azure cloud-node-manager configuration.
// reference: https://raw.githubusercontent.com/kubernetes-sigs/cloud-provider-azure/master/examples/out-of-tree/cloud-node-manager.yaml .
type CloudNodeManager struct{}

// Image returns the container image used to provide cloud-node-manager for the cloud-provider.
func (c *CloudNodeManager) Image() string {
	return cloudprovider.CloudNodeManagerImageAzure
}

// Path returns the path used by cloud-node-manager executable within the container image.
func (c *CloudNodeManager) Path() string {
	return "cloud-node-manager"
}

// ExtraArgs returns a list of arguments to append to the cloud-node-manager command.
func (c *CloudNodeManager) ExtraArgs() []string {
	return []string{
		"--wait-routes=true",
	}
}

// Supported is used to determine if cloud node manager is implemented for this cloud provider.
func (c *CloudNodeManager) Supported() bool {
	return true
}
