package kubectl

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	corev1 "k8s.io/api/core/v1"
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
	CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error
	AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string) error
	AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string) error
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
			return fmt.Errorf("kubectl apply of object %v/%v: %w", i, len(infos), err)
		}
	}

	return nil
}

// SetKubeconfig will store the kubeconfig to generate Clients using the clientGenerator later.
func (k *Kubectl) SetKubeconfig(kubeconfig []byte) {
	k.kubeconfig = kubeconfig
}

func (k *Kubectl) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	client, err := k.clientGenerator.NewClient(k.kubeconfig)
	if err != nil {
		return err
	}

	err = client.CreateConfigMap(ctx, configMap)
	if err != nil {
		return err
	}

	return nil
}

func (k *Kubectl) AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string) error {
	client, err := k.clientGenerator.NewClient(k.kubeconfig)
	if err != nil {
		return err
	}

	if err = client.AddTolerationsToDeployment(ctx, tolerations, name); err != nil {
		return err
	}

	return nil
}

func (k *Kubectl) AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string) error {
	client, err := k.clientGenerator.NewClient(k.kubeconfig)
	if err != nil {
		return err
	}

	if err = client.AddNodeSelectorsToDeployment(ctx, selectors, name); err != nil {
		return err
	}

	return nil
}
