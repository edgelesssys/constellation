/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestK8sCompliantHostname(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
		wantErr  bool
	}{
		"no change": {
			input:    "test",
			expected: "test",
		},
		"uppercase": {
			input:    "TEST",
			expected: "test",
		},
		"underscore": {
			input:    "test_node",
			expected: "test-node",
		},
		"empty": {
			input:    "",
			expected: "",
			wantErr:  true,
		},
		"error": {
			input:    "test_node_",
			expected: "",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := k8sCompliantHostname(tc.input)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.expected, actual)
		})
	}
}
