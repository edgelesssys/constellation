/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client is a kubernetes client.
type Client struct {
	client    *kubernetes.Clientset
	dynClient dynamic.Interface
}

// New creates a new kubernetes client.
func New() (*Client, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{client: clientset, dynClient: dynClient}, nil
}

// GetComponents returns the components of the cluster.
func (c *Client) GetComponents(ctx context.Context, configMapName string) (versions.ComponentVersions, error) {
	componentsRaw, err := c.getConfigMapData(ctx, configMapName, constants.ComponentsListKey)
	if err != nil {
		return versions.ComponentVersions{}, fmt.Errorf("failed to get components: %w", err)
	}
	var components versions.ComponentVersions
	if err := json.Unmarshal([]byte(componentsRaw), &components); err != nil {
		return versions.ComponentVersions{}, fmt.Errorf("failed to unmarshal components %s: %w", componentsRaw, err)
	}
	return components, nil
}

func (c *Client) getConfigMapData(ctx context.Context, name, key string) (string, error) {
	cm, err := c.client.CoreV1().ConfigMaps("kube-system").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get configmap: %w", err)
	}

	return cm.Data[key], nil
}

// CreateConfigMap creates the provided configmap.
func (c *Client) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	_, err := c.client.CoreV1().ConfigMaps(configMap.ObjectMeta.Namespace).Create(ctx, &configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}
	return nil
}

// AddNodeToJoiningNodes adds the provided node as a joining node CRD.
func (c *Client) AddNodeToJoiningNodes(ctx context.Context, nodeName string, componentsHash string) error {
	joiningNodeResource := schema.GroupVersionResource{Group: "update.edgeless.systems", Version: "v1alpha1", Resource: "joiningnodes"}

	joiningNode := &unstructured.Unstructured{}
	joiningNode.SetUnstructuredContent(map[string]any{
		"apiVersion": "update.edgeless.systems/v1alpha1",
		"kind":       "JoiningNode",
		"metadata": map[string]any{
			"name": nodeName,
		},
		"spec": map[string]any{
			"name":           nodeName,
			"componentshash": componentsHash,
		},
	})

	_, err := c.dynClient.Resource(joiningNodeResource).Apply(ctx, joiningNode.GetName(), joiningNode, metav1.ApplyOptions{FieldManager: "join-service"})
	if err != nil {
		return fmt.Errorf("failed to create joining node: %w", err)
	}
	return nil
}

// AddReferenceToK8sVersionConfigMap adds a reference to the provided configmap to the k8s version configmap.
func (c *Client) AddReferenceToK8sVersionConfigMap(ctx context.Context, k8sVersionsConfigMapName string, componentsConfigMapName string) error {
	cm, err := c.client.CoreV1().ConfigMaps("kube-system").Get(ctx, k8sVersionsConfigMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configmap: %w", err)
	}
	cm.Data[constants.K8sComponentsFieldName] = componentsConfigMapName
	_, err = c.client.CoreV1().ConfigMaps("kube-system").Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap: %w", err)
	}
	return nil
}
