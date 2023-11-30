/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/license"
)

// An Applier handles applying a specific configuration to a Constellation cluster.
// In Particular, this involves Initialization and Upgrading of the cluster.
type Applier struct {
	log            debugLog
	licenseChecker *license.Checker
}

type debugLog interface {
	Debugf(format string, args ...any)
}

// NewApplier creates a new Applier.
func NewApplier(log debugLog) *Applier {
	return &Applier{
		log:            log,
		licenseChecker: license.NewChecker(license.NewClient()),
	}
}

// CheckLicense checks the given Constellation license with the license server
// and returns the allowed quota for the license.
func (a *Applier) CheckLicense(ctx context.Context, csp cloudprovider.Provider, licenseID string) (int, error) {
	a.log.Debugf("Contacting license server for license '%s'", licenseID)
	quotaResp, err := a.licenseChecker.CheckLicense(ctx, csp, licenseID)
	if err != nil {
		return 0, fmt.Errorf("checking license: %w", err)
	}
	a.log.Debugf("Got response from license server for license '%s'", licenseID)

	return quotaResp.Quota, nil
}
