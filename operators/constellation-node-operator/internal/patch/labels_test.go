/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var labelsTestObject = corev1.Node{
	TypeMeta: v1.TypeMeta{
		Kind:       "Node",
		APIVersion: "v1",
	},
	ObjectMeta: v1.ObjectMeta{
		ResourceVersion: "0",
	},
}

func TestSetLabels(t *testing.T) {
	testCases := map[string]struct {
		oldLabels map[string]string
		newLabels map[string]string
		wantPatch []byte
	}{
		"empty patch only contains resource version": {
			wantPatch: []byte(`{"metadata":{"resourceVersion":"0"}}`),
		},
		"patch on node without existing labels": {
			newLabels: map[string]string{"key": "value"},
			wantPatch: []byte(`{"metadata":{"labels":{"key":"value"},"resourceVersion":"0"}}`),
		},
		"patch on node with same existing labels": {
			oldLabels: map[string]string{"key": "value"},
			newLabels: map[string]string{"key": "value"},
			wantPatch: []byte(`{"metadata":{"resourceVersion":"0"}}`),
		},
		"patch on node with same key but different value": {
			oldLabels: map[string]string{"key": "oldvalue"},
			newLabels: map[string]string{"key": "newvalue"},
			wantPatch: []byte(`{"metadata":{"labels":{"key":"newvalue"},"resourceVersion":"0"}}`),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			original := labelsTestObject.DeepCopy()
			original.SetLabels(tc.oldLabels)
			patched := original.DeepCopy()
			patch := SetLabels(original, patched, tc.newLabels)
			data, err := patch.Data(patched)
			assert.NoError(err)
			assert.Equal(tc.wantPatch, data)
		})
	}
}
