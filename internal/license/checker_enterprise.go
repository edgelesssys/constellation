//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

type Checker struct {
	quotaChecker QuotaChecker
}

func NewChecker(quotaChecker QuotaChecker) *Checker {
	return &Checker{
		quotaChecker: quotaChecker,
	}
}

// CheckLicense contacts the license server to fetch quota information for the given license.
func (c *Checker) CheckLicense(ctx context.Context, provider cloudprovider.Provider, licenseID string) (QuotaCheckResponse, error) {
	return c.quotaChecker.QuotaCheck(ctx, QuotaCheckRequest{
		License:  licenseID,
		Action:   Init,
		Provider: provider.String(),
	})
}
