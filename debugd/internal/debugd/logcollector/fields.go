/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logcollector

// InfoFields are the fields that are allowed in the info map
// under the prefix "logcollect.".
func InfoFields() (string, map[string]struct{}) {
	return "logcollect.", map[string]struct{}{
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
		"github.e2e-test-provider": {},
		"deployment-type":          {}, // deployment type, e.g. "debugd", "k8s"
	}
}
