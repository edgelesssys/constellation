/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package kubecmd provides functions to interact with a Kubernetes cluster to the CLI.
The package should be used for:

  - Fetching status information about the cluster
  - Creating, deleting, or migrating resources not managed by Helm

The package should not be used for anything that doesn't just require the Kubernetes API.
For example, Terraform and Helm actions should not be accessed by this package.
*/
package kubecmd

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func newClient(kubeconfigPath string) (kubernetes.Interface, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up kubernetes client: %w", err)
	}
	return kubeClient, nil
}

type stableClient struct {
	client kubernetes.Interface
}

// GetConfigMap returns a ConfigMap given it's name.
func (c *stableClient) GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateConfigMap updates the given ConfigMap.
func (c *stableClient) UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
}

// CreateConfigMap creates the given ConfigMap.
func (c *stableClient) CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return c.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Create(ctx, configMap, metav1.CreateOptions{})
}

// KubernetesVersion returns the Kubernetes version of the cluster.
func (c *stableClient) KubernetesVersion() (string, error) {
	serverVersion, err := c.client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return serverVersion.GitVersion, nil
}

func (c *stableClient) GetNodes(ctx context.Context) ([]corev1.Node, error) {
	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	return nodes.Items, nil
}
