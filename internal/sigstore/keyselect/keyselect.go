/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package keyselect is used to select the correct public key for signature verification.
// The content of keyselect must be kept separate from internal/sigstore because keyselect relies on internal/api/versionsapi.
// Since internal/api relies on internal/sigstore, we need to separate the functions to avoid import cycles.
package keyselect

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// CosignPublicKeyForVersion returns the public key for the given version.
func CosignPublicKeyForVersion(ver versionsapi.Version) ([]byte, error) {
	if err := ver.Validate(); err != nil {
		return nil, fmt.Errorf("selecting public key: invalid version %s: %w", ver.ShortPath(), err)
	}
	if ver.Ref() == versionsapi.ReleaseRef && ver.Stream() == "stable" {
		return []byte(constants.CosignPublicKeyReleases), nil
	}
	return []byte(constants.CosignPublicKeyDev), nil
}
