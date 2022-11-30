/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logcollector

// InfoFields are the fields that are allowed in the info map
// under the prefix "logcollect.".
func InfoFields() (string, map[string]struct{}) {
	return "logcollect.", map[string]struct{}{
		"admin": {}, // the name of the person running the cdbg command
	}
}
