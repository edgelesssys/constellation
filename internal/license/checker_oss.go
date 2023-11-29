//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// Checker checks the Constellation license.
type Checker struct{}

// NewChecker creates a new Checker.
func NewChecker(QuotaChecker) *Checker {
	return &Checker{}
}

// CheckLicense is a no-op for open source version of Constellation.
func (c *Checker) CheckLicense(ctx context.Context, provider cloudprovider.Provider, licenseID string) (QuotaCheckResponse, error) {
	return QuotaCheckResponse{}, nil
}
