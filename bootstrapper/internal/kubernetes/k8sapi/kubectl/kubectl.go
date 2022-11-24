/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubectl

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// ErrKubeconfigNotSet is the error value returned by Kubectl.Apply when SetKubeconfig was not called first.
var ErrKubeconfigNotSet = errors.New("kubeconfig not set")

// Kubectl implements acts like the Kubernetes "kubectl" tool.
type Kubectl struct {
	kubernetes.Interface
	apiextensionClient apiextensionsclientv1.ApiextensionsV1Interface
	builder            *resource.Builder
}

// New returns an empty Kubectl client. Need to call Initialize before usable.
func New() *Kubectl {
	return &Kubectl{}
}

// Initialize sets sets all required fields so the Kubectl client can be used.
func (k *Kubectl) Initialize(kubeconfig []byte) error {
	clientConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating k8s client config from kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating k8s client from kubeconfig: %w", err)
	}
	k.Interface = clientset

	apiextensionClient, err := apiextensionsclientv1.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating api extension client from kubeconfig: %w", err)
	}
	k.apiextensionClient = apiextensionClient

	restClientGetter, err := newRESTClientGetter(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating k8s RESTClientGetter from kubeconfig: %w", err)
	}
	k.builder = resource.NewBuilder(restClientGetter).Unstructured()

	return nil
}

// CreateConfigMap creates the provided configmap.
func (k *Kubectl) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	_, err := k.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Create(ctx, &configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// ListAllNamespaces returns all namespaces in the cluster.
func (k *Kubectl) ListAllNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	return k.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}

// AddTolerationsToDeployment adds [K8s tolerations] to the deployment, identified
// by name and namespace.
//
// [K8s tolerations]: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
func (k *Kubectl) AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error {
	deployments := k.AppsV1().Deployments(namespace)

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

// AddNodeSelectorsToDeployment adds [K8s selectors] to the deployment, identified
// by name and namespace.
//
// [K8s selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
func (k *Kubectl) AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error {
	deployments := k.AppsV1().Deployments(namespace)

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

// WaitForCRDs waits for a list of CRDs to be established.
func (k *Kubectl) WaitForCRDs(ctx context.Context, crds []string) error {
	for _, crd := range crds {
		err := k.WaitForCRD(ctx, crd)
		if err != nil {
			return err
		}
	}
	return nil
}

// WaitForCRD waits for the given CRD to be established.
func (k *Kubectl) WaitForCRD(ctx context.Context, crd string) error {
	watcher, err := k.apiextensionClient.CustomResourceDefinitions().Watch(ctx, metav1.ListOptions{
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
