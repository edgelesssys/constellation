/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
