/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kubewaiter is used to wait for the Kubernetes API to be available.
package kubewaiter

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/edgelesssys/constellation/v2/internal/retry"
)

// KubernetesClient is an interface for the Kubernetes client.
// It is used to check if the Kubernetes API is available.
type KubernetesClient interface {
	ListAllNamespaces(ctx context.Context) (*corev1.NamespaceList, error)
}

// CloudKubeAPIWaiter waits for the Kubernetes API to be available.
type CloudKubeAPIWaiter struct{}

// Wait waits for the Kubernetes API to be available.
// Note that the kubernetesClient must have the kubeconfig already set.
func (w *CloudKubeAPIWaiter) Wait(ctx context.Context, kubernetesClient KubernetesClient) error {
	funcAlwaysRetriable := func(_ error) bool { return true }

	doer := &kubeDoer{kubeClient: kubernetesClient}
	retrier := retry.NewIntervalRetrier(doer, 5*time.Second, funcAlwaysRetriable)
	if err := retrier.Do(ctx); err != nil {
		return fmt.Errorf("waiting for Kubernetes API: %w", err)
	}
	return nil
}

type kubeDoer struct {
	kubeClient KubernetesClient
}

func (d *kubeDoer) Do(ctx context.Context) error {
	_, err := d.kubeClient.ListAllNamespaces(ctx)
	return err
}
