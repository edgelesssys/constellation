/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"cloud.google.com/go/compute/apiv1/computepb"
)

// getMetadataByKey returns the value of the metadata key in the given metadata.
func getMetadataByKey(metadata *computepb.Metadata, key string) string {
	if metadata == nil {
		return ""
	}
	for _, item := range metadata.Items {
		if item.Key == nil || item.Value == nil {
			continue
		}
		if *item.Key == key {
			return *item.Value
		}
	}
	return ""
}
