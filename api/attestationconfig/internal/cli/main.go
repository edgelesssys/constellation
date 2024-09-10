/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
This package provides a CLI to interact with the Attestationconfig API, a sub API of the Resource API.

You can execute an e2e test by running: `bazel run //internal/api/attestationconfig:configapi_e2e_test`.
The CLI is used in the CI pipeline. Manual actions that change the bucket's data shouldn't be necessary.
The reporter CLI caches the observed version values in a dedicated caching directory and derives the latest API version from it.
Any version update is then pushed to the API.
*/
package main

import (
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

const (
	awsRegion           = "eu-central-1"
	awsBucket           = "cdn-constellation-backend"
	distributionID      = constants.CDNDefaultDistributionID
	envCosignPwd        = "COSIGN_PASSWORD"
	envCosignPrivateKey = "COSIGN_PRIVATE_KEY"
	// versionWindowSize defines the number of versions to be considered for the latest version.
	// Through our weekly e2e tests, each week 2 versions are uploaded:
	// One from a stable release, and one from a debug image.
	// A window size of 6 ensures we update only after a version has been "stable" for 3 weeks.
	versionWindowSize = 6
)

var (
	// Cosign credentials.
	cosignPwd  string
	privateKey string
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// newRootCmd creates the root command.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Short: "CLI to interact with the attestationconfig API",
		Long:  "CLI to interact with the attestationconfig API. Allows uploading new TCB versions, deleting specific versions and deleting all versions. Uploaded objects are signed with cosign.",
	}
	rootCmd.PersistentFlags().StringP("region", "r", awsRegion, "region of the targeted bucket.")
	rootCmd.PersistentFlags().StringP("bucket", "b", awsBucket, "bucket targeted by all operations.")
	rootCmd.PersistentFlags().Bool("testing", false, "upload to S3 test bucket.")

	rootCmd.AddCommand(newUploadCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newCompareCmd())

	return rootCmd
}

type apiConfig struct {
	url             string
	distribution    string
	cosignPublicKey string
}

func getAPIEnvironment(testing bool) apiConfig {
	if testing {
		return apiConfig{url: "https://d33dzgxuwsgbpw.cloudfront.net", distribution: "ETZGUP1CWRC2P", cosignPublicKey: constants.CosignPublicKeyDev}
	}
	return apiConfig{url: constants.CDNRepositoryURL, distribution: constants.CDNDefaultDistributionID, cosignPublicKey: constants.CosignPublicKeyReleases}
}
