/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kubernetes interacts with the Kubernetes API to update an fetch objects related to joining nodes.
package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
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
func (c *Client) GetComponents(ctx context.Context, configMapName string) (components.Components, error) {
	componentsRaw, err := c.GetConfigMapData(ctx, configMapName, constants.ComponentsListKey)
	if err != nil {
		return components.Components{}, fmt.Errorf("failed to get components: %w", err)
	}
	var clusterComponents components.Components
	if err := json.Unmarshal([]byte(componentsRaw), &clusterComponents); err != nil {
		return components.Components{}, fmt.Errorf("failed to unmarshal components %s: %w", componentsRaw, err)
	}
	return clusterComponents, nil
}

// GetConfigMapData returns the data for the given key in the configmap with the given name.
func (c *Client) GetConfigMapData(ctx context.Context, name, key string) (string, error) {
	cm, err := c.client.CoreV1().ConfigMaps("kube-system").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get configmap: %w", err)
	}

	return cm.Data[key], nil
}

// GetK8sComponentsRefFromNodeVersionCRD returns the K8sComponentsRef from the node version CRD.
func (c *Client) GetK8sComponentsRefFromNodeVersionCRD(ctx context.Context, nodeName string) (string, error) {
	nodeVersionResource := schema.GroupVersionResource{Group: "update.edgeless.systems", Version: "v1alpha1", Resource: "nodeversions"}
	nodeVersion, err := c.dynClient.Resource(nodeVersionResource).Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get node version: %w", err)
	}
	// Extract K8sComponentsRef from nodeVersion.
	k8sComponentsRef, found, err := unstructured.NestedString(nodeVersion.Object, "spec", "kubernetesComponentsReference")
	if err != nil {
		return "", fmt.Errorf("failed to get K8sComponentsRef from node version: %w", err)
	}
	if !found {
		return "", fmt.Errorf("kubernetesComponentsReference not found in node version")
	}
	return k8sComponentsRef, nil
}

// AddNodeToJoiningNodes adds the provided node as a joining node CRD.
func (c *Client) AddNodeToJoiningNodes(ctx context.Context, nodeName string, componentsReference string, isControlPlane bool) error {
	joiningNode := &unstructured.Unstructured{}

	compliantNodeName, err := k8sCompliantHostname(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get k8s compliant hostname: %w", err)
	}

	// JoiningNodes referencing a worker node are named after the worker node.
	// JoiningNodes referencing the control-plane node are named "control-plane".
	objectMetadataName := compliantNodeName
	deadline := metav1.NewTime(time.Now().Add(48 * time.Hour))
	if isControlPlane {
		objectMetadataName = "control-plane"
		deadline = metav1.NewTime(time.Now().Add(10 * time.Minute))
	}

	joiningNode.SetUnstructuredContent(map[string]any{
		"apiVersion": "update.edgeless.systems/v1alpha1",
		"kind":       "JoiningNode",
		"metadata": map[string]any{
			"name": objectMetadataName,
		},
		"spec": map[string]any{
			"name":                compliantNodeName,
			"componentsreference": componentsReference,
			"iscontrolplane":      isControlPlane,
			"deadline":            deadline,
		},
	})
	if isControlPlane {
		return c.addControlPlaneToJoiningNodes(ctx, joiningNode)
	}
	return c.addWorkerToJoiningNodes(ctx, joiningNode)
}

func (c *Client) addControlPlaneToJoiningNodes(ctx context.Context, joiningNode *unstructured.Unstructured) error {
	joiningNodeResource := schema.GroupVersionResource{Group: "update.edgeless.systems", Version: "v1alpha1", Resource: "joiningnodes"}
	_, err := c.dynClient.Resource(joiningNodeResource).Create(ctx, joiningNode, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create joining control-plane node, maybe another node is already joining: %w", err)
	}
	return nil
}

func (c *Client) addWorkerToJoiningNodes(ctx context.Context, joiningNode *unstructured.Unstructured) error {
	joiningNodeResource := schema.GroupVersionResource{Group: "update.edgeless.systems", Version: "v1alpha1", Resource: "joiningnodes"}
	_, err := c.dynClient.Resource(joiningNodeResource).Apply(ctx, joiningNode.GetName(), joiningNode, metav1.ApplyOptions{FieldManager: "join-service"})
	if err != nil {
		return fmt.Errorf("failed to create joining node: %w", err)
	}
	return nil
}

var validHostnameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

// k8sCompliantHostname transforms a hostname to an RFC 1123 compliant, lowercase subdomain as required by Kubernetes node names.
// Only a simple heuristic is used for now (to lowercase, replace underscores).
func k8sCompliantHostname(in string) (string, error) {
	hostname := strings.ToLower(in)
	hostname = strings.ReplaceAll(hostname, "_", "-")
	if !validHostnameRegex.MatchString(hostname) {
		return "", fmt.Errorf("failed to generate a Kubernetes compliant hostname for %s", in)
	}
	return hostname, nil
}

// CreateConfigMap creates a configmap in the kube-system namespace with the provided name and data.
func (c *Client) CreateConfigMap(ctx context.Context, name string, data map[string]string) error {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
		},
		Data: data,
	}
	_, err := c.client.CoreV1().ConfigMaps("kube-system").Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}
	return nil
}
