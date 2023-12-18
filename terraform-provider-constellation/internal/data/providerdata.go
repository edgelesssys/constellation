/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package data

import "github.com/edgelesssys/constellation/v2/internal/semver"

// ProviderData is the data that get's passed down from the provider
// configuration to the resources and data sources.
type ProviderData struct {
	Version semver.Semver
}
