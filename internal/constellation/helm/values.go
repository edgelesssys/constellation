/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

var controlPlaneNodeSelector = map[string]any{"node-role.kubernetes.io/control-plane": ""}

var controlPlaneTolerations = []map[string]any{
	{
		"key":      "node-role.kubernetes.io/control-plane",
		"effect":   "NoSchedule",
		"operator": "Exists",
	},
	{
		"key":      "node-role.kubernetes.io/master",
		"effect":   "NoSchedule",
		"operator": "Exists",
	},
}
