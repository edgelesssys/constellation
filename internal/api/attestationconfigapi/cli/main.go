/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
This package provides a CLI to interact with the Attestationconfig API, a sub API of the Resource API.

You can execute an e2e test by running: `bazel run //internal/api/attestationconfigapi:configapi_e2e_test`.
The CLI is used in the CI pipeline. Manual actions that change the bucket's data shouldn't be necessary.
The reporter CLI caches the observed version values in a dedicated caching directory and derives the latest API version from it.
Any version update is then pushed to the API.
Notice that there is no synchronization on API operations. // TODO(elchead): what does this mean?
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

const (
	awsRegion           = "eu-central-1"
	awsBucket           = "cdn-constellation-backend"
	distributionID      = constants.CDNDefaultDistributionID
	envCosignPwd        = "COSIGN_PASSWORD"
	envCosignPrivateKey = "COSIGN_PRIVATE_KEY"
	// versionWindowSize defines the number of versions to be considered for the latest version. Each week 5 versions are uploaded for each node of the verify cluster.
	versionWindowSize = 15
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
		Use:   "COSIGN_PASSWORD=$CPW COSIGN_PRIVATE_KEY=$CKEY upload --version-file $FILE",
		Short: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api.",

		Long: fmt.Sprintf("The CLI uploads an observed version number specific to the azure-sev-snp attestation variant to a cache directory. The CLI then determines the lowest version within the cache-window present in the cache and writes that value to the config api if necessary. "+
			"Please authenticate with AWS through your preferred method (e.g. environment variables, CLI)"+
			"to be able to upload to S3. Set the %s and %s environment variables to authenticate with cosign.",
			envCosignPrivateKey, envCosignPwd,
		),
		PreRunE: envCheck,
		RunE:    runCmd,
	}
	rootCmd.Flags().StringP("maa-claims-path", "t", "", "File path to a json file containing the MAA claims.")
	rootCmd.Flags().StringP("upload-date", "d", "", "upload a version with this date as version name.")
	rootCmd.Flags().BoolP("force", "f", false, "Use force to manually push a new latest version."+
		" The version gets saved to the cache but the version selection logic is skipped.")
	rootCmd.Flags().IntP("cache-window-size", "s", versionWindowSize, "Number of versions to be considered for the latest version.")
	rootCmd.PersistentFlags().StringP("region", "r", awsRegion, "region of the targeted bucket.")
	rootCmd.PersistentFlags().StringP("bucket", "b", awsBucket, "bucket targeted by all operations.")
	rootCmd.PersistentFlags().StringP("distribution", "i", distributionID, "cloudflare distribution used.")
	must(rootCmd.MarkFlagRequired("maa-claims-path"))
	rootCmd.AddCommand(newDeleteCmd())
	return rootCmd
}

func envCheck(_ *cobra.Command, _ []string) error {
	if os.Getenv(envCosignPrivateKey) == "" || os.Getenv(envCosignPwd) == "" {
		return fmt.Errorf("please set both %s and %s environment variables", envCosignPrivateKey, envCosignPwd)
	}
	cosignPwd = os.Getenv(envCosignPwd)
	privateKey = os.Getenv(envCosignPrivateKey)
	return nil
}

func runCmd(cmd *cobra.Command, _ []string) (retErr error) {
	ctx := cmd.Context()
	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")

	flags, err := parseCliFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing cli flags: %w", err)
	}

	cfg := staticupload.Config{
		Bucket:         flags.bucket,
		Region:         flags.region,
		DistributionID: flags.distribution,
	}

	log.Infof("Reading MAA claims from file: %s", flags.maaFilePath)
	maaClaimsBytes, err := os.ReadFile(flags.maaFilePath)
	if err != nil {
		return fmt.Errorf("reading MAA claims file: %w", err)
	}
	var maaTCB maaTokenTCBClaims
	if err = json.Unmarshal(maaClaimsBytes, &maaTCB); err != nil {
		return fmt.Errorf("unmarshalling MAA claims file: %w", err)
	}
	inputVersion := maaTCB.ToAzureSEVSNPVersion()
	log.Infof("Input version: %+v", inputVersion)

	client, clientClose, err := attestationconfigapi.NewClient(ctx, cfg,
		[]byte(cosignPwd), []byte(privateKey), false, flags.cacheWindowSize, log)
	defer func() {
		err := clientClose(cmd.Context())
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to invalidate cache: %w", err))
		}
	}()

	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	url := "https://d33dzgxuwsgbpw.cloudfront.net"
	latestAPIVersionAPI, err := attestationconfigapi.NewFetcherWithCustomCDNAndCosignKey(url, constants.CosignPublicKeyDev).FetchAzureSEVSNPVersionLatest(ctx)
	if err != nil {
		if errors.Is(err, attestationconfigapi.ErrNoVersionsFound) {
			log.Infof("No versions found in API, but assuming that we are uploading the first version.")
		} else {
			return fmt.Errorf("fetching latest version: %w", err)
		}
	}
	latestAPIVersion := latestAPIVersionAPI.AzureSEVSNPVersion
	if err := client.UploadAzureSEVSNPVersionLatest(ctx, inputVersion, latestAPIVersion, flags.uploadDate, flags.force); err != nil {
		if errors.Is(err, attestationconfigapi.ErrNoNewerVersion) {
			log.Infof("Input version: %+v is not newer than latest API version: %+v", inputVersion, latestAPIVersion)
			return nil
		}
		return fmt.Errorf("updating latest version: %w", err)
	}
	return nil
}

type cliFlags struct {
	maaFilePath     string
	uploadDate      time.Time
	region          string
	bucket          string
	distribution    string
	force           bool
	cacheWindowSize int
}

func parseCliFlags(cmd *cobra.Command) (cliFlags, error) {
	maaFilePath, err := cmd.Flags().GetString("maa-claims-path")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting maa claims path: %w", err)
	}

	dateStr, err := cmd.Flags().GetString("upload-date")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting upload date: %w", err)
	}
	uploadDate := time.Now()
	if dateStr != "" {
		uploadDate, err = time.Parse(attestationconfigapi.VersionFormat, dateStr)
		if err != nil {
			return cliFlags{}, fmt.Errorf("parsing date: %w", err)
		}
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting bucket: %w", err)
	}

	distribution, err := cmd.Flags().GetString("distribution")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting distribution: %w", err)
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting force: %w", err)
	}

	cacheWindowSize, err := cmd.Flags().GetInt("cache-window-size")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting cache window size: %w", err)
	}
	return cliFlags{
		maaFilePath:     maaFilePath,
		uploadDate:      uploadDate,
		region:          region,
		bucket:          bucket,
		distribution:    distribution,
		force:           force,
		cacheWindowSize: cacheWindowSize,
	}, nil
}

// maaTokenTCBClaims describes the TCB information in a MAA token.
type maaTokenTCBClaims struct {
	IsolationTEE struct {
		TEESvn        uint8 `json:"x-ms-sevsnpvm-tee-svn"`
		SNPFwSvn      uint8 `json:"x-ms-sevsnpvm-snpfw-svn"`
		MicrocodeSvn  uint8 `json:"x-ms-sevsnpvm-microcode-svn"`
		BootloaderSvn uint8 `json:"x-ms-sevsnpvm-bootloader-svn"`
	} `json:"x-ms-isolation-tee"`
}

func (c maaTokenTCBClaims) ToAzureSEVSNPVersion() attestationconfigapi.AzureSEVSNPVersion {
	return attestationconfigapi.AzureSEVSNPVersion{
		TEE:        c.IsolationTEE.TEESvn,
		SNP:        c.IsolationTEE.SNPFwSvn,
		Microcode:  c.IsolationTEE.MicrocodeSvn,
		Bootloader: c.IsolationTEE.BootloaderSvn,
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
