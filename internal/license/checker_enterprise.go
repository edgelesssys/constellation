//go:build enterprise

package license

import (
	"context"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
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
func (c *Checker) CheckLicense(ctx context.Context, printer func(string, ...any)) error {
	licenseID, err := FromFile(c.fileHandler, constants.LicenseFilename)
	if err != nil {
		printer("Unable to find license file. Assuming community license.\n")
		licenseID = CommunityLicense
	} else {
		printer("Constellation license found!\n")
	}
	quotaResp, err := c.quotaChecker.QuotaCheck(ctx, QuotaCheckRequest{
		License: licenseID,
		Action:  Init,
	})
	if err != nil {
		printer("Unable to contact license server.\n")
		printer("Please keep your vCPU quota in mind.\n")
	} else {
		printer("Please keep your vCPU quota (%d) in mind.\n", quotaResp.Quota)
	}
	return nil
}
