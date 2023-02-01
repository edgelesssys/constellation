/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConfigMaps represent a list of k8s ConfigMap.
type ConfigMaps []*corev1.ConfigMap

// Marshal marshals config maps into multiple YAML documents.
func (s ConfigMaps) Marshal() ([]byte, error) {
	objects := make([]runtime.Object, len(s))
	for i := range s {
		objects[i] = s[i]
	}
	return MarshalK8SResourcesList(objects)
}

// ConstructK8sComponentsCM creates a k8s-components config map for the given components.
func ConstructK8sComponentsCM(components components.Components, clusterVersion string) (corev1.ConfigMap, error) {
	componentsMarshalled, err := json.Marshal(components)
	if err != nil {
		return corev1.ConfigMap{}, fmt.Errorf("marshalling component versions: %w", err)
	}

	componentsHash := components.GetHash()
	componentConfigMapName := fmt.Sprintf("k8s-components-%s", strings.ReplaceAll(componentsHash, ":", "-"))

	return corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		Immutable: toPtr(true),
		ObjectMeta: metav1.ObjectMeta{
			Name:      componentConfigMapName,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			constants.ComponentsListKey:   string(componentsMarshalled),
			constants.K8sVersionFieldName: clusterVersion,
		},
	}, nil
}

func toPtr[T any](v T) *T {
	return &v
}
