package qemu

import "github.com/edgelesssys/constellation/internal/versions"

// CloudNodeManager holds the QEMU cloud-node-manager configuration.
type CloudNodeManager struct{}

// Image returns the container image used to provide cloud-node-manager for the cloud-provider.
// Not used on QEMU.
func (c *CloudNodeManager) Image(k8sVersion versions.ValidK8sVersion) (string, error) {
	return "", nil
}

// Path returns the path used by cloud-node-manager executable within the container image.
// Not used on QEMU.
func (c *CloudNodeManager) Path() string {
	return ""
}

// ExtraArgs returns a list of arguments to append to the cloud-node-manager command.
// Not used on QEMU.
func (c *CloudNodeManager) ExtraArgs() []string {
	return []string{}
}

// Supported is used to determine if cloud node manager is implemented for this cloud provider.
func (c *CloudNodeManager) Supported() bool {
	return false
}
