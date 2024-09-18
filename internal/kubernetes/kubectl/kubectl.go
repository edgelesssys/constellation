/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package kubectl provides a kubectl-like interface for Kubernetes.
Functions defined here should not make use of [os/exec].
*/
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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// NewUninitialized returns an empty Kubectl client.
// Initialize needs to be called before the client is usable.
func NewUninitialized() *Kubectl {
	return &Kubectl{}
}

// NewFromConfig returns a Kubectl client using the given kubeconfig.
func NewFromConfig(kubeconfig []byte) (*Kubectl, error) {
	k := NewUninitialized()
	if err := k.Initialize(kubeconfig); err != nil {
		return nil, err
	}
	return k, nil
}

// Initialize sets sets all required fields so the Kubectl client can be used.
func (k *Kubectl) Initialize(kubeconfig []byte) error {
	clientConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating k8s client config from kubeconfig: %w", err)
	}
	if err := k.initialize(clientConfig); err != nil {
		return fmt.Errorf("initializing kubectl: %w", err)
	}
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

// ListCRDs retrieves all custom resource definitions currently installed in the cluster.
func (k *Kubectl) ListCRDs(ctx context.Context) ([]apiextensionsv1.CustomResourceDefinition, error) {
	crds, err := k.apiextensionClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing CRDs: %w", err)
	}

	return crds.Items, nil
}

// ListCRs retrieves all objects for a given CRD.
func (k *Kubectl) ListCRs(ctx context.Context, gvr schema.GroupVersionResource) ([]unstructured.Unstructured, error) {
	unstructuredList, err := k.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing CRs for GroupVersionResource %+v: %w", gvr, err)
	}

	return unstructuredList.Items, nil
}

// GetCR retrieves a Custom Resource given it's name and group version resource.
func (k *Kubectl) GetCR(ctx context.Context, gvr schema.GroupVersionResource, name string) (*unstructured.Unstructured, error) {
	return k.dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
}

// UpdateCR updates a Custom Resource given it's and group version resource.
func (k *Kubectl) UpdateCR(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return k.dynamicClient.Resource(gvr).Update(ctx, obj, metav1.UpdateOptions{})
}

// CreateConfigMap creates the provided configmap.
func (k *Kubectl) CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error {
	_, err := k.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// GetConfigMap returns a ConfigMap given it's name and namespace.
func (k *Kubectl) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	return k.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateConfigMap updates the given ConfigMap.
func (k *Kubectl) UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return k.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Update(ctx, configMap, metav1.UpdateOptions{})
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

// KubernetesVersion returns the Kubernetes version of the cluster.
func (k *Kubectl) KubernetesVersion() (string, error) {
	serverVersion, err := k.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return serverVersion.GitVersion, nil
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

// PatchFirstNodePodCIDR patches the firstNodePodCIDR of the first control-plane node for Cilium.
func (k *Kubectl) PatchFirstNodePodCIDR(ctx context.Context, firstNodePodCIDR string) error {
	selector := labels.Set{"node-role.kubernetes.io/control-plane": ""}.AsSelector()
	controlPlaneList, err := k.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return err
	}
	if len(controlPlaneList.Items) != 1 {
		return fmt.Errorf("expected 1 control-plane node, got %d", len(controlPlaneList.Items))
	}
	nodeName := controlPlaneList.Items[0].Name
	// Update the node's spec
	_, err = k.CoreV1().Nodes().Patch(context.Background(), nodeName, types.MergePatchType, []byte(fmt.Sprintf(`{"spec":{"podCIDR":"%s"}}`, firstNodePodCIDR)), metav1.PatchOptions{})
	return err
}

// EnforceCoreDNSSpread adds a pod anti-affinity to the CoreDNS deployment to ensure that
// CoreDNS pods are spread across nodes.
func (k *Kubectl) EnforceCoreDNSSpread(ctx context.Context) error {
	// allow CoreDNS Pods to run on uninitialized nodes, which is required by cloud-controller-manager
	tolerationSeconds := int64(10)
	tolerations := []corev1.Toleration{
		{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOpExists,
			Effect:            corev1.TaintEffectNoExecute,
			TolerationSeconds: &tolerationSeconds,
		},
	}

	deployments := k.AppsV1().Deployments("kube-system")
	// retry resource update if an error occurs
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := deployments.Get(ctx, "coredns", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get Deployment to add toleration: %w", err)
		}

		result.Spec.Template.Spec.Tolerations = append(result.Spec.Template.Spec.Tolerations, tolerations...)

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

func (k *Kubectl) initialize(clientConfig *rest.Config) error {
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating k8s client from kubeconfig: %w", err)
	}
	k.Interface = clientset

	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating unstructured client: %w", err)
	}
	k.dynamicClient = dynamicClient

	apiextensionClient, err := apiextensionsclientv1.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("creating api extension client from kubeconfig: %w", err)
	}
	k.apiextensionClient = apiextensionClient

	return nil
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
