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
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
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
	rootCmd := &cobra.Command{}
	rootCmd.PersistentFlags().StringP("region", "r", awsRegion, "region of the targeted bucket.")
	rootCmd.PersistentFlags().StringP("bucket", "b", awsBucket, "bucket targeted by all operations.")
	rootCmd.PersistentFlags().Bool("testing", false, "upload to S3 test bucket.")

	rootCmd.AddCommand(newUploadCmd())
	rootCmd.AddCommand(newDeleteCmd())

	return rootCmd
}

func newUploadCmd() *cobra.Command {
	uploadCmd := &cobra.Command{
		Use:   "upload {azure|aws} {snp-report|guest-firmware} <path>",
		Short: "Upload an object to the attestationconfig API",

		Long: fmt.Sprintf("Upload a new object to the attestationconfig API. For snp-reports the new object is added to a cache folder first."+
			"The CLI then determines the lowest version within the cache-window present in the cache and writes that value to the config api if necessary. "+
			"For guest-firmware objects the object is added to the API directly. "+
			"Please authenticate with AWS through your preferred method (e.g. environment variables, CLI)"+
			"to be able to upload to S3. Set the %s and %s environment variables to authenticate with cosign.",
			envCosignPrivateKey, envCosignPwd,
		),

		Args:    cobra.MatchAll(cobra.ExactArgs(3), isCloudProvider(0), isValidKind(1)),
		PreRunE: envCheck,
		RunE:    runUpload,
	}
	uploadCmd.Flags().StringP("upload-date", "d", "", "upload a version with this date as version name.")
	uploadCmd.Flags().BoolP("force", "f", false, "Use force to manually push a new latest version."+
		" The version gets saved to the cache but the version selection logic is skipped.")
	uploadCmd.Flags().IntP("cache-window-size", "s", versionWindowSize, "Number of versions to be considered for the latest version.")

	return uploadCmd
}

func envCheck(_ *cobra.Command, _ []string) error {
	if os.Getenv(envCosignPrivateKey) == "" || os.Getenv(envCosignPwd) == "" {
		return fmt.Errorf("please set both %s and %s environment variables", envCosignPrivateKey, envCosignPwd)
	}
	cosignPwd = os.Getenv(envCosignPwd)
	privateKey = os.Getenv(envCosignPrivateKey)
	return nil
}

func runUpload(cmd *cobra.Command, args []string) (retErr error) {
	ctx := cmd.Context()
	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")

	log.Infof("%s", args)
	uploadCfg, err := newConfig(cmd, ([3]string)(args[:3]))
	if err != nil {
		return fmt.Errorf("parsing cli flags: %w", err)
	}

	client, clientClose, err := attestationconfigapi.NewClient(
		ctx,
		staticupload.Config{
			Bucket:         uploadCfg.bucket,
			Region:         uploadCfg.region,
			DistributionID: uploadCfg.distribution,
		},
		[]byte(cosignPwd),
		[]byte(privateKey),
		false,
		uploadCfg.cacheWindowSize,
		log)

	defer func() {
		err := clientClose(cmd.Context())
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to invalidate cache: %w", err))
		}
	}()

	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	switch uploadCfg.provider {
	case cloudprovider.AWS:
		return uploadAWS(ctx, client, uploadCfg, file.NewHandler(afero.NewOsFs()), log)
	case cloudprovider.Azure:
		return uploadAzure(ctx, client, uploadCfg, file.NewHandler(afero.NewOsFs()), log)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", uploadCfg.provider)
	}
}

type uploadConfig struct {
	provider        cloudprovider.Provider
	kind            objectKind
	path            string
	uploadDate      time.Time
	cosignPublicKey string
	region          string
	bucket          string
	distribution    string
	url             string
	force           bool
	cacheWindowSize int
}

func newConfig(cmd *cobra.Command, args [3]string) (uploadConfig, error) {
	dateStr, err := cmd.Flags().GetString("upload-date")
	if err != nil {
		return uploadConfig{}, fmt.Errorf("getting upload date: %w", err)
	}
	uploadDate := time.Now()
	if dateStr != "" {
		uploadDate, err = time.Parse(attestationconfigapi.VersionFormat, dateStr)
		if err != nil {
			return uploadConfig{}, fmt.Errorf("parsing date: %w", err)
		}
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return uploadConfig{}, fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return uploadConfig{}, fmt.Errorf("getting bucket: %w", err)
	}

	testing, err := cmd.Flags().GetBool("testing")
	if err != nil {
		return uploadConfig{}, fmt.Errorf("getting testing flag: %w", err)
	}
	apiCfg := getAPIEnvironment(testing)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return uploadConfig{}, fmt.Errorf("getting force: %w", err)
	}

	cacheWindowSize, err := cmd.Flags().GetInt("cache-window-size")
	if err != nil {
		return uploadConfig{}, fmt.Errorf("getting cache window size: %w", err)
	}

	provider := cloudprovider.FromString(args[0])
	kind := kindFromString(args[1])
	path := args[2]

	return uploadConfig{
		provider:        provider,
		kind:            kind,
		path:            path,
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
