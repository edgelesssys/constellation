/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Provides functionality to easily interact with the K8s API, which can be used
// from any e2e test.
package kubectl

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// New creates a new k8s client. The kube config file is expected to be set
// via environment variable KUBECONFIG or located at ./constellation-admin.conf.
func New() (*kubernetes.Clientset, error) {
	cfgPath := ""
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		cfgPath = envPath
		fmt.Printf("K8s config path empty. Using environment variable %s=%s.\n", "KUBECONFIG", envPath)
	} else {
		cfgPath = "constellation-admin.conf"
		fmt.Printf("K8s config path empty. Assuming '%s'.\n", cfgPath)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return k8sClient, nil
}
