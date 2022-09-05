/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubectl

import "github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/kubectl/client"

// generator implements clientGenerator interface.
type generator struct{}

// NewClients generates a new client implementing the Client interface.
func (generator) NewClient(kubeconfig []byte) (Client, error) {
	return client.New(kubeconfig)
}
