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

// WaitForDS waits for a DaemonSet to become ready.
func WaitForDS(ctx context.Context, kubeconfigPath, namespace, name string, log debugLog) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Wait for the DaemonSet to become ready
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context expired before DaemonSet %q became ready", name)
		default:
			ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, v1.GetOptions{})
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
func RestartDS(kubeconfigPath, namespace, name string) error {
	// Load the Kubernetes configuration from file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Get the current DaemonSet
	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return err
	}

	// Update the DaemonSet's template
	ds.Spec.Template.ObjectMeta.Annotations["restartTimestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	_, err = clientset.AppsV1().DaemonSets(namespace).Update(context.Background(), ds, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
