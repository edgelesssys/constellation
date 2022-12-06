/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	testCases := map[string]struct {
		vals      map[string]any
		extraVals map[string]any
		expected  map[string]any
	}{
		"equal": {
			vals: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
					"key2": "bar",
				},
			},
			extraVals: map[string]any{
				"join-service": map[string]any{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
			expected: map[string]any{
				"join-service": map[string]any{
					"key1":      "foo",
					"key2":      "bar",
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
		},
		"missing join-service extraVals": {
			vals: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
					"key2": "bar",
				},
			},
			extraVals: map[string]any{
				"extraKey1": "extraFoo",
				"extraKey2": "extraBar",
			},
			expected: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
					"key2": "bar",
				},
				"extraKey1": "extraFoo",
				"extraKey2": "extraBar",
			},
		},
		"missing join-service vals": {
			vals: map[string]any{
				"key1": "foo",
				"key2": "bar",
			},
			extraVals: map[string]any{
				"join-service": map[string]any{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
			expected: map[string]any{
				"key1": "foo",
				"key2": "bar",
				"join-service": map[string]any{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
		},
		"key collision": {
			vals: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
				},
			},
			extraVals: map[string]any{
				"join-service": map[string]any{
					"key1": "bar",
				},
			},
			expected: map[string]any{
				"join-service": map[string]any{
					"key1": "bar",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			newVals := MergeMaps(tc.vals, tc.extraVals)
			assert.Equal(tc.expected, newVals)
		})
	}
}
