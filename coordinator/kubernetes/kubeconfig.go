package kubernetes

import (
	"fmt"

	"github.com/spf13/afero"
)

const kubeconfigPath = "/etc/kubernetes/admin.conf"

// KubeconfigReader implements ConfigReader.
type KubeconfigReader struct {
	fs afero.Afero
}

// ReadKubeconfig reads the Kubeconfig from disk.
func (r KubeconfigReader) ReadKubeconfig() ([]byte, error) {
	kubeconfig, err := r.fs.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("reading gce config: %w", err)
	}
	return kubeconfig, nil
}
