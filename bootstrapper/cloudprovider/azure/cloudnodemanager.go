package azure

import (
	"fmt"

	"github.com/edgelesssys/constellation/internal/versions"
)

// CloudNodeManager holds the Azure cloud-node-manager configuration.
// reference: https://raw.githubusercontent.com/kubernetes-sigs/cloud-provider-azure/master/examples/out-of-tree/cloud-node-manager.yaml .
type CloudNodeManager struct{}

// Image returns the container image used to provide cloud-node-manager for the cloud-provider.
func (c *CloudNodeManager) Image(k8sVersion string) (string, error) {
	if !versions.IsSupportedK8sVersion(k8sVersion) {
		return "", fmt.Errorf("received unsupported k8sVersion: %s", k8sVersion)
	}
	return versions.VersionConfigs[k8sVersion].CloudControllerManagerImageAzure, nil
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
