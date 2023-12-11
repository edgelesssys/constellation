/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package components

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// Components is a list of Kubernetes components.
type Components []*Component

// GetHash returns the hash over all component hashes.
func (c Components) GetHash() string {
	sha := sha256.New()
	for _, component := range c {
		sha.Write([]byte(component.Hash))
	}

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}

// GetKubeadmComponent returns the kubeadm component.
func (c Components) GetKubeadmComponent() (*Component, error) {
	for _, component := range c {
		if strings.Contains(component.GetUrl(), "kubeadm") {
			return component, nil
		}
	}
	return nil, errors.New("kubeadm component not found")
}
