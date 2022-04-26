package kubectl

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"k8s.io/cli-runtime/pkg/resource"
)

// ErrKubeconfigNotSet is the error value returned by Kubectl.Apply when SetKubeconfig was not called first.
var ErrKubeconfigNotSet = errors.New("kubeconfig not set")

// Client wraps marshable k8s resources into resource.Info fields and applies them in a cluster.
type Client interface {
	// ApplyOneObject applies a k8s resource similar to kubectl apply.
	ApplyOneObject(info *resource.Info, forceConflicts bool) error
	// GetObjects converts resources into prepared info fields for use in ApplyOneObject.
	GetObjects(resources resources.Marshaler) ([]*resource.Info, error)
}

// clientGenerator can generate new clients from a kubeconfig.
type clientGenerator interface {
	NewClient(kubeconfig []byte) (Client, error)
}

// Kubectl implements kubernetes.Apply interface and acts like the Kubernetes "kubectl" tool.
type Kubectl struct {
	clientGenerator
	kubeconfig []byte
}

// New creates a new kubectl using the real clientGenerator.
func New() *Kubectl {
	return &Kubectl{
		clientGenerator: &generator{},
	}
}

// Apply will apply the given resources using server-side-apply.
func (k *Kubectl) Apply(resources resources.Marshaler, forceConflicts bool) error {
	if k.kubeconfig == nil {
		return ErrKubeconfigNotSet
	}
	client, err := k.clientGenerator.NewClient(k.kubeconfig)
	if err != nil {
		return err
	}
	// convert marshaler object into []*resource.info
	infos, err := client.GetObjects(resources)
	if err != nil {
		return err
	}

	// apply each object, one by one
	for i, resource := range infos {
		if err := client.ApplyOneObject(resource, forceConflicts); err != nil {
			return fmt.Errorf("kubectl apply of object %v/%v failed: %w", i, len(infos), err)
		}
	}

	return nil
}

// SetKubeconfig will store the kubeconfig to generate Clients using the clientGenerator later.
func (k *Kubectl) SetKubeconfig(kubeconfig []byte) {
	k.kubeconfig = kubeconfig
}
