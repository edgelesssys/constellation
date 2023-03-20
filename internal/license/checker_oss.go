//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// Checker checks the Constellation license.
type Checker struct{}

// NewChecker creates a new Checker.
func NewChecker(_ QuotaChecker, _ file.Handler) *Checker {
	return &Checker{}
}

// CheckLicense is a no-op for open source version of Constellation.
func (c *Checker) CheckLicense(_ context.Context, _ cloudprovider.Provider, _ config.ProviderConfig, _ func(string, ...any)) error {
	return nil
}
