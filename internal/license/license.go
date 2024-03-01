/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package license provides functions to check a user's Constellation license.
package license

// Action performed by Constellation.
type Action string

const (
	// CommunityLicense is used by everyone who has not bought an enterprise license.
	CommunityLicense = "00000000-0000-0000-0000-000000000000"
	// MarketplaceLicense is used by everyone who uses a marketplace image.
	MarketplaceLicense = "11111111-1111-1111-1111-111111111111"

	// Init action denotes the initialization of a Constellation cluster.
	Init Action = "init"
	// Apply action denotes an update of a Constellation cluster.
	// It is used after a cluster has already been initialized once.
	Apply Action = "apply"
)
