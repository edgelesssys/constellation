/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigMaps(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	configMaps := ConfigMaps{
		&k8s.ConfigMap{
			TypeMeta: v1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			Data: map[string]string{"key": "value1"},
		},
		&k8s.ConfigMap{
			TypeMeta: v1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			Data: map[string]string{"key": "value2"},
		},
	}
	data, err := configMaps.Marshal()
	require.NoError(err)

	assert.Equal(`apiVersion: v1
data:
  key: value1
kind: ConfigMap
metadata:
  creationTimestamp: null
---
apiVersion: v1
data:
  key: value2
kind: ConfigMap
metadata:
  creationTimestamp: null
`, string(data))
}
