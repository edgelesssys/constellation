/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package client

// AutoscalingCloudProvider returns the cloud-provider name as used by k8s cluster-autoscaler.
func (c *Client) AutoscalingCloudProvider() string {
	return "azure"
}
