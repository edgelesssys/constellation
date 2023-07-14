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
		"github.actor":       {},
		"github.workflow":    {},
		"github.run-id":      {},
		"github.run-attempt": {},
		"github.ref-name":    {},
		"github.sha":         {},
		"github.runner-os":   {},
	}
}
