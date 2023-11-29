//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
)

// checkLicenseFile reads the local license file and checks it's quota
// with the license server. If no license file is present or if errors
// occur during the check, the user is informed and the community license
// is used. It is a no-op in the open source version of Constellation.
func (a *applyCmd) checkLicenseFile(*cobra.Command, cloudprovider.Provider) {}
