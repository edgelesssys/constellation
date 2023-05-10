/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

// AutoscalingCloudProvider returns the cloud-provider name as used by k8s cluster-autoscaler.
func (c *Client) AutoscalingCloudProvider() string {
	return "aws"
}
