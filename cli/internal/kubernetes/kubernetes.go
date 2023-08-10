/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package kubernetes provides functions to interact with a Kubernetes cluster to the CLI.
The package should be used for:

  - Fetching status information about the cluster
  - Creating, deleting, or migrating resources not managed by Helm

The package should not be used for anything that doesn't just require the Kubernetes API.
For example, Terraform and Helm actions should not be accessed by this package.
*/
package kubernetes

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

// StableInterface is an interface to interact with stable resources.
type StableInterface interface {
	GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
	UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	KubernetesVersion() (string, error)
}

// NewStableClient returns a new StableClient.
func NewStableClient(kubeconfigPath string) (StableInterface, error) {
	client, err := newClient(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return &stableClient{client}, nil
}

type stableClient struct {
	client kubernetes.Interface
}

// GetConfigMap returns a ConfigMap given it's name.
func (u *stableClient) GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateConfigMap updates the given ConfigMap.
func (u *stableClient) UpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
}

// CreateConfigMap creates the given ConfigMap.
func (u *stableClient) CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return u.client.CoreV1().ConfigMaps(constants.ConstellationNamespace).Create(ctx, configMap, metav1.CreateOptions{})
}

// KubernetesVersion returns the Kubernetes version of the cluster.
func (u *stableClient) KubernetesVersion() (string, error) {
	serverVersion, err := u.client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return serverVersion.GitVersion, nil
}
