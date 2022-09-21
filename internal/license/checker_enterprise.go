//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"context"
	"errors"
	"io/fs"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

type Checker struct {
	quotaChecker QuotaChecker
	fileHandler  file.Handler
}

func NewChecker(quotaChecker QuotaChecker, fileHandler file.Handler) *Checker {
	return &Checker{
		quotaChecker: quotaChecker,
		fileHandler:  fileHandler,
	}
}

// CheckLicense tries to read the license file and contact license server
// to fetch quota information.
// If no license file is found, community license is assumed.
func (c *Checker) CheckLicense(ctx context.Context, provider cloudprovider.Provider, providerCfg config.ProviderConfig, printer func(string, ...any)) error {
	licenseID, err := FromFile(c.fileHandler, constants.LicenseFilename)
	if errors.Is(err, fs.ErrNotExist) {
		printer("Unable to find license file. Assuming community license.\n")
		licenseID = CommunityLicense
	} else if err != nil {
		printer("Error: %v\nContinuing with community license.\n", err)
		licenseID = CommunityLicense
	} else {
		printer("Constellation license found!\n")
	}
	providerStr := provider.String()
	if provider == cloudprovider.Azure {
		if *providerCfg.Azure.ConfidentialVM {
			providerStr = "azure-cvm"
		} else {
			providerStr = "azure-tl"
		}
	}

	quotaResp, err := c.quotaChecker.QuotaCheck(ctx, QuotaCheckRequest{
		License:  licenseID,
		Action:   Init,
		Provider: providerStr,
	})
	if err != nil {
		printer("Unable to contact license server.\n")
		printer("Please keep your vCPU quota in mind.\n")
	} else if licenseID == CommunityLicense {
		printer("You can use Constellation to create services for internal consumption.\n")
		printer("For details, see https://docs.edgeless.systems/constellation/overview/license\n")
	} else {
		printer("Please keep your vCPU quota (%d) in mind.\n", quotaResp.Quota)
	}
	return nil
}
