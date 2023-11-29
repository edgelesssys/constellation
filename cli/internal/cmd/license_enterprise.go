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
func (a *applyCmd) checkLicenseFile(cmd *cobra.Command, csp cloudprovider.Provider) {
	var licenseID string
	a.log.Debugf("Running license check")

	readBytes, err := a.fileHandler.Read(constants.LicenseFilename)
	if errors.Is(err, fs.ErrNotExist) {
		cmd.Printf("Using community license.\n")
		licenseID = license.CommunityLicense
	} else if err != nil {
		cmd.Printf("Error: %v\nContinuing with community license.\n", err)
		licenseID = license.CommunityLicense
	} else {
		cmd.Printf("Constellation license found!\n")
		licenseID, err = license.FromBytes(readBytes)
		if err != nil {
			cmd.Printf("Error: %v\nContinuing with community license.\n", err)
			licenseID = license.CommunityLicense
		}
	}

	quota, err := a.applier.CheckLicense(cmd.Context(), csp, licenseID)
	if err != nil {
		cmd.Printf("Unable to contact license server.\n")
		cmd.Printf("Please keep your vCPU quota in mind.\n")
	} else if licenseID == license.CommunityLicense {
		cmd.Printf("For details, see https://docs.edgeless.systems/constellation/overview/license\n")
	} else {
		cmd.Printf("Please keep your vCPU quota (%d) in mind.\n", quota)
	}

	a.log.Debugf("Checked license")
}
