/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package migration contains outdated configuration formats and their migration functions.
package migration

import (
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/file"
)

// V2ToV3 converts an existing v2 config to a v3 config.
func V2ToV3(_ string, _ file.Handler) error {
	// TODO(malt3): add migration from v3 to v4
	return errors.New("not implemented")
}
