/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package logcollector

// THIS FILE IS A DUPLICATE OF hack/logcollector/fields/fields.go

import (
	"fmt"
	"strings"
)

var (
	// DebugdLogcollectPrefix is the prefix for all OpenSearch fields specified by the user when starting through debugd.
	DebugdLogcollectPrefix = "logcollect."
	// AllowedFields are the fields that are allowed to be used in the logcollection.
	AllowedFields = map[string]struct{}{
		"admin":            {}, // name of the person running the cdbg command
		"is_debug_cluster": {}, // whether the cluster is a debug cluster
		// GitHub workflow information, see https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
		"github.actor":            {},
		"github.workflow":         {},
		"github.run-id":           {},
		"github.run-attempt":      {},
		"github.ref-name":         {},
		"github.sha":              {},
		"github.runner-os":        {},
		"github.e2e-test-payload": {},
		"github.is-debug-cluster": {},
		// cloud provider used in e2e test. If deployed with debugd, this is a duplicate as its also
		// available in the metadata. If deployed through K8s in e2e tests with a stable image, this
		// is where the cloud provider is saved in.
		"github.e2e-test-provider":  {},
		"github.ref-stream":         {},
		"github.kubernetes-version": {},
		"github.cluster-creation":   {},
		"deployment-type":           {}, // deployment type, e.g. "debugd", "k8s"
	}
)

// FieldsFromMap returns new Fields from the given map.
func FieldsFromMap(m map[string]string) Fields {
	return Fields(m)
}

// Fields are the OpenSearch fields that are associated with a log message.
type Fields map[string]string

// Extend adds the fields from other to f and returns the result.
func (f Fields) Extend(other Fields) Fields {
	for k, v := range other {
		f[k] = v
	}
	return f
}

// Check checks whether all the fields in f are allowed. For fields that are prefixed with the debugd logcollect prefix are
// only the subkeys are checked.
func (f Fields) Check() error {
	for k := range f {
		if !strings.HasPrefix(k, DebugdLogcollectPrefix) {
			continue
		}
		subkey := strings.TrimPrefix(k, DebugdLogcollectPrefix)

		if _, ok := AllowedFields[subkey]; !ok {
			return fmt.Errorf("invalid subkey %q for info key %q", subkey, fmt.Sprintf("%s%s", DebugdLogcollectPrefix, k))
		}
	}

	return nil
}
