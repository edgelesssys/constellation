//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"io/fs"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/spf13/cobra"
)

// checkLicenseFile reads the local license file and checks it's quota
// with the license server. If no license file is present or if errors
// occur during the check, the user is informed and the community license
// is used. It is a no-op in the open source version of Constellation.
func (a *applyCmd) checkLicenseFile(cmd *cobra.Command, csp cloudprovider.Provider, useMarketplaceImage bool) {
	var licenseID string
	a.log.Debug("Running license check")

	readBytes, err := a.fileHandler.Read(constants.LicenseFilename)
	switch {
	case useMarketplaceImage:
		cmd.Println("Using marketplace image billing.")
		licenseID = license.MarketplaceLicense
	case errors.Is(err, fs.ErrNotExist):
		cmd.Println("Using community license.")
		licenseID = license.CommunityLicense
	case err != nil:
		cmd.Printf("Error: %v\nContinuing with community license.\n", err)
		licenseID = license.CommunityLicense
	default:
		cmd.Printf("Constellation license found!\n")
		licenseID, err = license.FromBytes(readBytes)
		if err != nil {
			cmd.Printf("Error: %v\nContinuing with community license.\n", err)
			licenseID = license.CommunityLicense
		}
	}

	quota, err := a.applier.CheckLicense(cmd.Context(), csp, !a.flags.skipPhases.contains(skipInitPhase), licenseID)
	if err != nil && !useMarketplaceImage {
		cmd.Printf("Unable to contact license server.\n")
		cmd.Printf("Please keep your vCPU quota in mind.\n")
	} else if licenseID == license.MarketplaceLicense {
		// Do nothing. Billing is handled by the marketplace.
	} else if licenseID == license.CommunityLicense {
		cmd.Printf("For details, see https://docs.edgeless.systems/constellation/overview/license\n")
	} else {
		cmd.Printf("Please keep your vCPU quota (%d) in mind.\n", quota)
	}

	a.log.Debug("Checked license")
}
