package client

import (
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
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
