/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var annotationsTestObject = corev1.Node{
	TypeMeta: v1.TypeMeta{
		Kind:       "Node",
		APIVersion: "v1",
	},
	ObjectMeta: v1.ObjectMeta{
		ResourceVersion: "0",
	},
}

func TestSetAnnotations(t *testing.T) {
	testCases := map[string]struct {
		oldAnnotations map[string]string
		newAnnotations map[string]string
		wantPatch      []byte
	}{
		"empty patch only contains resource version": {
			wantPatch: []byte(`{"metadata":{"resourceVersion":"0"}}`),
		},
		"patch on node without existing annotations": {
			newAnnotations: map[string]string{"key": "value"},
			wantPatch:      []byte(`{"metadata":{"annotations":{"key":"value"},"resourceVersion":"0"}}`),
		},
		"patch on node with same existing annotations": {
			oldAnnotations: map[string]string{"key": "value"},
			newAnnotations: map[string]string{"key": "value"},
			wantPatch:      []byte(`{"metadata":{"resourceVersion":"0"}}`),
		},
		"patch on node with same key but different value": {
			oldAnnotations: map[string]string{"key": "oldvalue"},
			newAnnotations: map[string]string{"key": "newvalue"},
			wantPatch:      []byte(`{"metadata":{"annotations":{"key":"newvalue"},"resourceVersion":"0"}}`),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			original := annotationsTestObject.DeepCopy()
			original.SetAnnotations(tc.oldAnnotations)
			patched := original.DeepCopy()
			patch := SetAnnotations(original, patched, tc.newAnnotations)
			data, err := patch.Data(patched)
			assert.NoError(err)
			assert.Equal(tc.wantPatch, data)
		})
	}
}

func TestUnsetAnnotations(t *testing.T) {
	testCases := map[string]struct {
		oldAnnotations map[string]string
		annotationKeys []string
		wantPatch      []byte
	}{
		"empty patch only contains resource version": {
			wantPatch: []byte(`{"metadata":{"resourceVersion":"0"}}`),
		},
		"patch on node without existing annotations": {
			annotationKeys: []string{"key"},
			wantPatch:      []byte(`{"metadata":{"resourceVersion":"0"}}`),
		},
		"patch on node with existing annotations": {
			oldAnnotations: map[string]string{"key": "value"},
			annotationKeys: []string{"key"},
			wantPatch:      []byte(`{"metadata":{"annotations":null,"resourceVersion":"0"}}`),
		},
		"patch on node with existing annotations delete one of multiple": {
			oldAnnotations: map[string]string{"key": "value", "otherkey": "othervalue"},
			annotationKeys: []string{"key"},
			wantPatch:      []byte(`{"metadata":{"annotations":{"key":null},"resourceVersion":"0"}}`),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			original := annotationsTestObject.DeepCopy()
			original.SetAnnotations(tc.oldAnnotations)
			patched := original.DeepCopy()
			patch := UnsetAnnotations(original, patched, tc.annotationKeys)
			data, err := patch.Data(patched)
			assert.NoError(err)
			assert.Equal(tc.wantPatch, data)
		})
	}
}
