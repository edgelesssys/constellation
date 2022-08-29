package client

import (
	"bytes"
	"context"
	"fmt"

	kubernetesshared "github.com/edgelesssys/constellation/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

const fieldManager = "constellation-bootstrapper"

// Client implements k8sapi.Client interface and talks to the Kubernetes API.
type Client struct {
	clientset          kubernetes.Interface
	apiextensionClient apiextensionsclientv1.ApiextensionsV1Interface
	builder            *resource.Builder
}

// New creates a new Client, talking to the real k8s API.
func New(config []byte) (*Client, error) {
	clientConfig, err := clientcmd.RESTConfigFromKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating k8s client config from kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("creating k8s client from kubeconfig: %w", err)
	}

	apiextensionClient, err := apiextensionsclientv1.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("creating api extension client from kubeconfig: %w", err)
	}

	restClientGetter, err := newRESTClientGetter(config)
	if err != nil {
		return nil, fmt.Errorf("creating k8s RESTClientGetter from kubeconfig: %w", err)
	}
	builder := resource.NewBuilder(restClientGetter).Unstructured()

	return &Client{clientset: clientset, apiextensionClient: apiextensionClient, builder: builder}, nil
}

// ApplyOneObject uses server-side apply to send unstructured JSON blobs to the server and let it handle the core logic.
func (c *Client) ApplyOneObject(info *resource.Info, forceConflicts bool) error {
	// helper can be used to patch k8s resources using server-side-apply.
	helper := resource.NewHelper(info.Client, info.Mapping).
		WithFieldManager(fieldManager)

	// server-side-apply uses unstructured JSON instead of strict typing on the client side.
	data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, info.Object)
	if err != nil {
		return fmt.Errorf("preparing resource for server-side apply: encoding of resource: %w", err)
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
		return fmt.Errorf("applying object %v using server-side apply: %w", info, err)
	}

	return info.Refresh(obj, true)
}

// GetObjects tries to marshal the resources into []*resource.Info using a resource.Builder.
func (c *Client) GetObjects(resources kubernetesshared.Marshaler) ([]*resource.Info, error) {
	// convert our resource struct into YAML
	data, err := resources.Marshal()
	if err != nil {
		return nil, fmt.Errorf("converting resources to YAML: %w", err)
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

// CreateConfigMap creates the given ConfigMap.
func (c *Client) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	_, err := c.clientset.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Create(ctx, &configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error {
	deployments := c.clientset.AppsV1().Deployments(namespace)

	// retry resource update if an error occurs
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := deployments.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get Deployment to add toleration: %v", err)
		}

		result.Spec.Template.Spec.Tolerations = append(result.Spec.Template.Spec.Tolerations, tolerations...)
		if _, err = deployments.Update(ctx, result, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error {
	deployments := c.clientset.AppsV1().Deployments(namespace)

	// retry resource update if an error occurs
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := deployments.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get Deployment to add node selector: %v", err)
		}

		for k, v := range selectors {
			result.Spec.Template.Spec.NodeSelector[k] = v
		}

		if _, err = deployments.Update(ctx, result, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// WaitForCRD waits for the given CRD to be established.
func (c *Client) WaitForCRD(ctx context.Context, crd string) error {
	watcher, err := c.apiextensionClient.CustomResourceDefinitions().Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", crd),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			crd := event.Object.(*apiextensionsv1.CustomResourceDefinition)
			if crdHasCondition(crd.Status.Conditions, apiextensionsv1.Established) {
				return nil
			}
		case watch.Deleted:
			return fmt.Errorf("crd %q deleted", crd)
		case watch.Error:
			return fmt.Errorf("crd %q error: %v", crd, event.Object)
		}
	}
	return fmt.Errorf("crd %q not established", crd)
}

func crdHasCondition(conditions []apiextensionsv1.CustomResourceDefinitionCondition, conditionType apiextensionsv1.CustomResourceDefinitionConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType && condition.Status == apiextensionsv1.ConditionTrue {
			return true
		}
	}
	return false
}
