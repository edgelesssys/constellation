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

func TestSecrets(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	secrets := Secrets{
		&k8s.Secret{
			TypeMeta: v1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			Data: map[string][]byte{"key": []byte("value1")},
		},
		&k8s.Secret{
			TypeMeta: v1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			Data: map[string][]byte{"key": []byte("value2")},
		},
	}
	data, err := secrets.Marshal()
	require.NoError(err)

	assert.Equal(`apiVersion: v1
data:
  key: dmFsdWUx
kind: Secret
metadata:
  creationTimestamp: null
---
apiVersion: v1
data:
  key: dmFsdWUy
kind: Secret
metadata:
  creationTimestamp: null
`, string(data))
}
