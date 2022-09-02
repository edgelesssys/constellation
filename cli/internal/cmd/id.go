/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

type clusterIDsFile struct {
	ClusterID string `json:"clusterID,omitempty"`
	OwnerID   string `json:"ownerID,omitempty"`
	IP        string `json:"ip,omitempty"`
}
