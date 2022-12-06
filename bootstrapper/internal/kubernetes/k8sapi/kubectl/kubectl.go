/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubectl

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiextensionsclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// Kubectl implements functionality of the Kubernetes "kubectl" tool.
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

// AnnotateNode adds the provided annotations to the node, identified by name.
func (k *Kubectl) AnnotateNode(ctx context.Context, nodeName, annotationKey, annotationValue string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node, err := k.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if node.Annotations == nil {
			node.Annotations = map[string]string{}
		}
		node.Annotations[annotationKey] = annotationValue
		_, err = k.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		return err
	})
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
			return fmt.Errorf("failed to get Deployment to add toleration: %w", err)
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
			return fmt.Errorf("failed to get Deployment to add node selector: %w", err)
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
