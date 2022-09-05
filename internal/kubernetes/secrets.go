/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	k8s "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Secrets represent a list of k8s Secret.
type Secrets []*k8s.Secret

// Marshal marshals secrets into multiple YAML documents.
func (s Secrets) Marshal() ([]byte, error) {
	objects := make([]runtime.Object, len(s))
	for i := range s {
		objects[i] = s[i]
	}
	return MarshalK8SResourcesList(objects)
}
