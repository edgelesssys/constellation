/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kubectl provides a kubectl-like interface for Kubernetes.
// Functions defined here should not make use of [os/exec].
package kubectl

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/scale/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// Kubectl implements functionality of the Kubernetes "kubectl" tool.
type Kubectl struct {
	kubernetes.Interface
	dynamicClient      dynamic.Interface
	apiextensionClient apiextensionsclientv1.ApiextensionsV1Interface
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

	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating unstructed client: %w", err)
	}
	k.dynamicClient = dynamicClient

	apiextensionClient, err := apiextensionsclientv1.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating api extension client from kubeconfig: %w", err)
	}
	k.apiextensionClient = apiextensionClient

	return nil
}

// ApplyCRD updates the given CRD by parsing it, querying it's version from the cluster and finally updating it.
func (k *Kubectl) ApplyCRD(ctx context.Context, rawCRD []byte) error {
	crd, err := parseCRD(rawCRD)
	if err != nil {
		return fmt.Errorf("parsing crds: %w", err)
	}

	clusterCRD, err := k.apiextensionClient.CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting crd: %w", err)
	}
	crd.ResourceVersion = clusterCRD.ResourceVersion
	_, err = k.apiextensionClient.CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
	return err
}

// GetCRDs retrieves all custom resource definitions currently installed in the cluster.
func (k *Kubectl) GetCRDs(ctx context.Context) ([]apiextensionsv1.CustomResourceDefinition, error) {
	crds, err := k.apiextensionClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing CRDs: %w", err)
	}

	return crds.Items, nil
}

// GetCRs retrieves all objects for a given CRD.
func (k *Kubectl) GetCRs(ctx context.Context, gvr schema.GroupVersionResource) ([]unstructured.Unstructured, error) {
	crdClient := k.dynamicClient.Resource(gvr)
	unstructuredList, err := crdClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing CRDs for gvr %+v: %w", crdClient, err)
	}

	return unstructuredList.Items, nil
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

// GetNodes returns all nodes in the cluster.
func (k *Kubectl) GetNodes(ctx context.Context) ([]corev1.Node, error) {
	nodes, err := k.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	return nodes.Items, nil
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
	return err
}

// EnforceCoreDNSSpread adds a pod anti-affinity to the coredns deployment to ensure that
// coredns pods are spread across nodes.
func (k *Kubectl) EnforceCoreDNSSpread(ctx context.Context) error {
	deployments := k.AppsV1().Deployments("kube-system")
	// retry resource update if an error occurs
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := deployments.Get(ctx, "coredns", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get Deployment to add toleration: %w", err)
		}

		if result.Spec.Template.Spec.Affinity == nil {
			result.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		if result.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			result.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
		}
		result.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{}
		if result.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			result.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []corev1.PodAffinityTerm{}
		}

		result.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(result.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "k8s-app",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"kube-dns"},
						},
					},
				},
				TopologyKey: "kubernetes.io/hostname",
			})

		_, err = deployments.Update(ctx, result, metav1.UpdateOptions{})
		return err
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

// DeleteStorageClass deletes the storage class with the given name.
// TODO(daniel-weisse): Remove with v2.9.
func (k *Kubectl) DeleteStorageClass(ctx context.Context, name string) error {
	return k.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
}

// parseCRD takes a byte slice of data and tries to create a CustomResourceDefinition object from it.
func parseCRD(crdString []byte) (*v1.CustomResourceDefinition, error) {
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = v1.AddToScheme(sch)
	obj, groupVersionKind, err := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode(crdString, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding crd: %w", err)
	}
	if groupVersionKind.Kind == "CustomResourceDefinition" {
		return obj.(*v1.CustomResourceDefinition), nil
	}

	return nil, errors.New("parsed []byte, but did not find a CRD")
}
