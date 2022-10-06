/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

// ConstellationState is the state of a Constellation.
type ConstellationState struct {
	Name           string `json:"name,omitempty"`
	UID            string `json:"uid,omitempty"`
	CloudProvider  string `json:"cloudprovider,omitempty"`
	LoadBalancerIP string `json:"bootstrapperhost,omitempty"`
}
