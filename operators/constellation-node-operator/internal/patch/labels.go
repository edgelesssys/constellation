/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package patch

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetLabels creates a patch for a client.Object by merging labels with existing labels.
func SetLabels(original, patched client.Object, labels map[string]string) client.Patch {
	mergedLabels := patched.GetLabels()
	if mergedLabels == nil {
		mergedLabels = make(map[string]string, len(labels))
	}
	for labelKey, labelValue := range labels {
		mergedLabels[labelKey] = labelValue
	}
	patched.SetLabels(mergedLabels)
	return client.StrategicMergeFrom(original, client.MergeFromWithOptimisticLock{})
}
