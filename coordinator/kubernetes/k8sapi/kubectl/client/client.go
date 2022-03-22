package client

import (
	"bytes"
	"fmt"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const fieldManager = "constellation-coordinator"

// Client implements k8sapi.Client interface and talks to the kubernetes API.
type Client struct {
	clientset kubernetes.Interface
	builder   *resource.Builder
}

// New creates a new Client, talking to the real k8s API.
func New(config []byte) (*Client, error) {
	clientConfig, err := clientcmd.RESTConfigFromKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating k8s client config from kubeconfig failed: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("creating k8s client from kubeconfig failed: %w", err)
	}

	restClientGetter, err := newRESTClientGetter(config)
	if err != nil {
		return nil, fmt.Errorf("creating k8s RESTClientGetter from kubeconfig failed: %w", err)
	}
	builder := resource.NewBuilder(restClientGetter).Unstructured()

	return &Client{clientset: clientset, builder: builder}, nil
}

// ApplyOneObject uses server-side apply to send unstructured JSON blobs to the server and let it handle the core logic.
func (c *Client) ApplyOneObject(info *resource.Info, forceConflicts bool) error {
	// helper can be used to patch k8s resources using server-side-apply.
	helper := resource.NewHelper(info.Client, info.Mapping).
		WithFieldManager(fieldManager)

	// server-side-apply uses unstructured JSON instead of strict typing on the client side.
	data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, info.Object)
	if err != nil {
		return fmt.Errorf("preparing resource for server-side apply failed: encoding of resource failed: %w", err)
	}
	options := metav1.PatchOptions{
		Force: &forceConflicts,
	}
	obj, err := helper.Patch(
		info.Namespace,
		info.Name,
		types.ApplyPatchType,
		data,
		&options,
	)
	if err != nil {
		return fmt.Errorf("failed to apply object %v using server-side apply: %w", info, err)
	}

	return info.Refresh(obj, true)
}

// GetObjects tries to marshal the resources into []*resource.Info using a resource.Builder.
func (c *Client) GetObjects(resources resources.Marshaler) ([]*resource.Info, error) {
	// convert our resource struct into YAML
	data, err := resources.Marshal()
	if err != nil {
		return nil, fmt.Errorf("converting resources to YAML failed: %w", err)
	}
	// read into resource.Info using builder
	reader := bytes.NewReader(data)
	result := c.builder.
		ContinueOnError().
		NamespaceParam("default").
		DefaultNamespace().
		Stream(reader, "yaml").
		Flatten().
		Do()
	return result.Infos()
}
