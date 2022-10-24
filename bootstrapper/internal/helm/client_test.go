/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMe(t *testing.T) {
	testCases := map[string]struct {
		vals      map[string]interface{}
		extraVals map[string]interface{}
		expected  map[string]interface{}
	}{
		"equal": {
			vals: map[string]interface{}{
				"join-service": map[string]interface{}{
					"key1": "foo",
					"key2": "bar",
				},
			},
			extraVals: map[string]interface{}{
				"join-service": map[string]interface{}{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
			expected: map[string]interface{}{
				"join-service": map[string]interface{}{
					"key1":      "foo",
					"key2":      "bar",
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
		},
		"missing join-service extraVals": {
			vals: map[string]interface{}{
				"join-service": map[string]interface{}{
					"key1": "foo",
					"key2": "bar",
				},
			},
			extraVals: map[string]interface{}{
				"extraKey1": "extraFoo",
				"extraKey2": "extraBar",
			},
			expected: map[string]interface{}{
				"join-service": map[string]interface{}{
					"key1": "foo",
					"key2": "bar",
				},
				"extraKey1": "extraFoo",
				"extraKey2": "extraBar",
			},
		},
		"missing join-service vals": {
			vals: map[string]interface{}{
				"key1": "foo",
				"key2": "bar",
			},
			extraVals: map[string]interface{}{
				"join-service": map[string]interface{}{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
			expected: map[string]interface{}{
				"key1": "foo",
				"key2": "bar",
				"join-service": map[string]interface{}{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			newVals := mergeMaps(tc.vals, tc.extraVals)
			assert.Equal(tc.expected, newVals)
		})
	}
}
