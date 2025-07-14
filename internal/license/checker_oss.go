//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package license

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// Checker checks the Constellation license.
type Checker struct{}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{}
}

// CheckLicense is a no-op for open source version of Constellation.
func (c *Checker) CheckLicense(context.Context, cloudprovider.Provider, Action, string) (int, error) {
	return 0, nil
}
