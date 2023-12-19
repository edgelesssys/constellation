/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package license provides functions to check a user's Constellation license.
package license

const (
	// CommunityLicense is used by everyone who has not bought an enterprise license.
	CommunityLicense = "00000000-0000-0000-0000-000000000000"
	// InitRequest denotes that the client is checking their quota on cluster initialization.
	InitRequest = true
	// ApplyRequest denotes that the client is checking their quota on cluster update.
	ApplyRequest = false
)
