package gcp

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/spf13/afero"
)

// ConfigWriter abstracts writing of the /etc/gce.conf config file.
type ConfigWriter interface {
	// WriteGCEConf writes the config to the filesystem of the current instance.
	WriteGCEConf(config string) error
}

// CloudControllerManager holds the gcp cloud-controller-manager configuration.
type CloudControllerManager struct {
	writer ConfigWriter
}

// NewCCM creates a new CloudControllerManager.
func NewCCM() *CloudControllerManager {
	return &CloudControllerManager{
		writer: &Writer{fs: afero.Afero{Fs: afero.NewOsFs()}},
	}
}

// Image returns the container image used to provide cloud-controller-manager for the cloud-provider.
func (c *CloudControllerManager) Image() string {
	// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available
	return "ghcr.io/malt3/cloud-provider-gcp:latest"
}

// Path returns the path used by cloud-controller-manager executable within the container image.
func (c *CloudControllerManager) Path() string {
	return "/cloud-controller-manager"
}

// Name returns the cloud-provider name as used by k8s cloud-controller-manager (k8s.gcr.io/cloud-controller-manager).
func (c *CloudControllerManager) Name() string {
	return "gce"
}

// PrepareInstance is called on every instance before deploying the cloud-controller-manager.
// Allows for cloud-provider specific hooks.
func (c *CloudControllerManager) PrepareInstance(instance core.Instance, vpnIP string) error {
	// GCP CCM expects "/etc/gce.conf" to contain the GCP project-id and other configuration.
	// reference: https://github.com/kubernetes/cloud-provider-gcp/blob/master/cluster/gce/gci/configure-helper.sh#L791-L892
	var config strings.Builder
	config.WriteString("[global]\n")
	projectID, _, _, err := splitProviderID(instance.ProviderID)
	if err != nil {
		return fmt.Errorf("retrieving GCP project-id failed: %w", err)
	}
	config.WriteString(fmt.Sprintf("project-id = %s\n", projectID))
	config.WriteString("use-metadata-server = false\n")

	return c.writer.WriteGCEConf(config.String())
}

// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
func (c *CloudControllerManager) Supported() bool {
	return true
}
