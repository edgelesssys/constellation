/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sDsClient struct {
	clientset *kubernetes.Clientset
}

func newK8sCiliumHelper(kubeconfigPath string) (*k8sDsClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &k8sDsClient{clientset: clientset}, nil
}

func (h *k8sDsClient) PatchNode(ctx context.Context, podCIDR string) error {
	selector := labels.Set{"node-role.kubernetes.io/control-plane": ""}.AsSelector()
	controlPlaneList, err := h.clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return err
	}
	if len(controlPlaneList.Items) != 1 {
		return fmt.Errorf("expected 1 control-plane node, got %d", len(controlPlaneList.Items))
	}
	nodeName := controlPlaneList.Items[0].Name
	// Get the current node
	node, err := h.clientset.CoreV1().Nodes().Get(context.Background(), nodeName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("node %s not found", nodeName)
		}
		return err
	}
	// Update the node's spec
	node.Spec.PodCIDR = podCIDR
	_, err = h.clientset.CoreV1().Nodes().Patch(context.Background(), nodeName, types.MergePatchType, []byte(fmt.Sprintf(`{"spec":{"podCIDR":"%s"}}`, podCIDR)), v1.PatchOptions{})
	return err
}

// WaitForDS waits for a DaemonSet to become ready.
func (h *k8sDsClient) WaitForDS(ctx context.Context, namespace, name string, log debugLog) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context expired before DaemonSet %q became ready", name)
		default:
			ds, err := h.clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, v1.GetOptions{})
			if err != nil {
				return err
			}

			if ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
				log.Debugf("DaemonSet %s is ready\n", name)
				return nil
			}

			log.Debugf("Waiting for DaemonSet %s to become ready...\n", name)
			time.Sleep(10 * time.Second)
		}
	}
}

// RestartDS restarts all pods of a DaemonSet by updating its template.
func (h *k8sDsClient) RestartDS(namespace, name string) error {
	ds, err := h.clientset.AppsV1().DaemonSets(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return err
	}

	ds.Spec.Template.ObjectMeta.Annotations["restartTimestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	_, err = h.clientset.AppsV1().DaemonSets(namespace).Update(context.Background(), ds, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
