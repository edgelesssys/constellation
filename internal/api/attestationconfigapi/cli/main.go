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
*/
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/edgelesssys/constellation/v2/internal/verify"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	rootCmd.Flags().StringP("snp-report-path", "t", "", "File path to a file containing the Constellation verify output.")
	rootCmd.Flags().StringP("upload-date", "d", "", "upload a version with this date as version name.")
	rootCmd.Flags().BoolP("force", "f", false, "Use force to manually push a new latest version."+
		" The version gets saved to the cache but the version selection logic is skipped.")
	rootCmd.Flags().IntP("cache-window-size", "s", versionWindowSize, "Number of versions to be considered for the latest version.")
	rootCmd.PersistentFlags().StringP("region", "r", awsRegion, "region of the targeted bucket.")
	rootCmd.PersistentFlags().StringP("bucket", "b", awsBucket, "bucket targeted by all operations.")
	rootCmd.PersistentFlags().Bool("testing", false, "upload to S3 test bucket.")
	must(rootCmd.MarkFlagRequired("snp-report-path"))
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

	log.Infof("Reading SNP report from file: %s", flags.snpReportPath)

	fs := file.NewHandler(afero.NewOsFs())
	var report verify.Report
	if err := fs.ReadJSON(flags.snpReportPath, &report); err != nil {
		return fmt.Errorf("reading snp report: %w", err)
	}
	snpReport := report.SNPReport
	if !allEqual(snpReport.LaunchTCB, snpReport.CommittedTCB, snpReport.ReportedTCB) {
		return fmt.Errorf("TCB versions are not equal: \nLaunchTCB:%+v\nCommitted TCB:%+v\nReportedTCB:%+v",
			snpReport.LaunchTCB, snpReport.CommittedTCB, snpReport.ReportedTCB)
	}
	inputVersion := convertTCBVersionToAzureVersion(snpReport.LaunchTCB)
	log.Infof("Input report: %+v", inputVersion)

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

	latestAPIVersionAPI, err := attestationconfigapi.NewFetcherWithCustomCDNAndCosignKey(flags.url, flags.cosignPublicKey).FetchAzureSEVSNPVersionLatest(ctx)
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

func allEqual(args ...verify.TCBVersion) bool {
	if len(args) < 2 {
		return true
	}

	firstArg := args[0]
	for _, arg := range args[1:] {
		if arg != firstArg {
			return false
		}
	}

	return true
}

func convertTCBVersionToAzureVersion(tcb verify.TCBVersion) attestationconfigapi.AzureSEVSNPVersion {
	return attestationconfigapi.AzureSEVSNPVersion{
		Bootloader: tcb.Bootloader,
		TEE:        tcb.TEE,
		SNP:        tcb.SNP,
		Microcode:  tcb.Microcode,
	}
}

type config struct {
	snpReportPath   string
	uploadDate      time.Time
	cosignPublicKey string
	region          string
	bucket          string
	distribution    string
	url             string
	force           bool
	cacheWindowSize int
}

func parseCliFlags(cmd *cobra.Command) (config, error) {
	snpReportFilePath, err := cmd.Flags().GetString("snp-report-path")
	if err != nil {
		return config{}, fmt.Errorf("getting maa claims path: %w", err)
	}

	dateStr, err := cmd.Flags().GetString("upload-date")
	if err != nil {
		return config{}, fmt.Errorf("getting upload date: %w", err)
	}
	uploadDate := time.Now()
	if dateStr != "" {
		uploadDate, err = time.Parse(attestationconfigapi.VersionFormat, dateStr)
		if err != nil {
			return config{}, fmt.Errorf("parsing date: %w", err)
		}
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return config{}, fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return config{}, fmt.Errorf("getting bucket: %w", err)
	}

	testing, err := cmd.Flags().GetBool("testing")
	if err != nil {
		return config{}, fmt.Errorf("getting testing flag: %w", err)
	}
	apiCfg := getAPIEnvironment(testing)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return config{}, fmt.Errorf("getting force: %w", err)
	}

	cacheWindowSize, err := cmd.Flags().GetInt("cache-window-size")
	if err != nil {
		return config{}, fmt.Errorf("getting cache window size: %w", err)
	}
	return config{
		snpReportPath:   snpReportFilePath,
		uploadDate:      uploadDate,
		cosignPublicKey: apiCfg.cosignPublicKey,
		region:          region,
		bucket:          bucket,
		url:             apiCfg.url,
		distribution:    apiCfg.distribution,
		force:           force,
		cacheWindowSize: cacheWindowSize,
	}, nil
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

func must(err error) {
	if err != nil {
		panic(err)
	}
}
