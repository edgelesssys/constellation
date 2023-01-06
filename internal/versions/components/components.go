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

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
)

// Component is a Kubernetes component.
type Component struct {
	URL         string
	Hash        string
	InstallPath string
	Extract     bool
}

// Components is a list of Kubernetes components.
type Components []Component

// NewComponentsFromInitProto converts a protobuf KubernetesVersion to Components.
func NewComponentsFromInitProto(protoComponents []*initproto.KubernetesComponent) Components {
	components := Components{}
	for _, protoComponent := range protoComponents {
		if protoComponent == nil {
			continue
		}
		components = append(components, Component{URL: protoComponent.Url, Hash: protoComponent.Hash, InstallPath: protoComponent.InstallPath, Extract: protoComponent.Extract})
	}
	return components
}

// NewComponentsFromJoinProto converts a protobuf KubernetesVersion to Components.
func NewComponentsFromJoinProto(protoComponents []*joinproto.KubernetesComponent) Components {
	components := Components{}
	for _, protoComponent := range protoComponents {
		if protoComponent == nil {
			continue
		}
		components = append(components, Component{URL: protoComponent.Url, Hash: protoComponent.Hash, InstallPath: protoComponent.InstallPath, Extract: protoComponent.Extract})
	}
	return components
}

// ToInitProto converts Components to a protobuf KubernetesVersion.
func (c Components) ToInitProto() []*initproto.KubernetesComponent {
	protoComponents := []*initproto.KubernetesComponent{}
	for _, component := range c {
		protoComponents = append(protoComponents, &initproto.KubernetesComponent{Url: component.URL, Hash: component.Hash, InstallPath: component.InstallPath, Extract: component.Extract})
	}
	return protoComponents
}

// ToJoinProto converts Components to a protobuf KubernetesVersion.
func (c Components) ToJoinProto() []*joinproto.KubernetesComponent {
	protoComponents := []*joinproto.KubernetesComponent{}
	for _, component := range c {
		protoComponents = append(protoComponents, &joinproto.KubernetesComponent{Url: component.URL, Hash: component.Hash, InstallPath: component.InstallPath, Extract: component.Extract})
	}
	return protoComponents
}

// GetHash returns the hash over all component hashes.
func (c Components) GetHash() string {
	sha := sha256.New()
	for _, component := range c {
		sha.Write([]byte(component.Hash))
	}

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}

// GetKubeadmComponent returns the kubeadm component.
func (c Components) GetKubeadmComponent() (Component, error) {
	for _, component := range c {
		if strings.Contains(component.URL, "kubeadm") {
			return component, nil
		}
	}
	return Component{}, errors.New("kubeadm component not found")
}
