/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

type commonFlags struct {
	rawImage  string
	pki       string
	provider  cloudprovider.Provider
	variant   string
	version   versionsapi.Version
	timestamp time.Time
	region    string
	bucket    string
	out       string
	logLevel  zapcore.Level
}

func parseCommonFlags(cmd *cobra.Command) (commonFlags, error) {
	workspaceDir := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	rawImage, err := cmd.Flags().GetString("raw-image")
	if err != nil {
		return commonFlags{}, err
	}
	pki, err := cmd.Flags().GetString("pki")
	if err != nil {
		return commonFlags{}, err
	}
	if pki == "" {
		pki = filepath.Join(workspaceDir, "image/pki")
	}
	variant, err := cmd.Flags().GetString("variant")
	if err != nil {
		return commonFlags{}, err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return commonFlags{}, err
	}
	ver, err := versionsapi.NewVersionFromShortPath(version, versionsapi.VersionKindImage)
	if err != nil {
		return commonFlags{}, err
	}
	timestamp, err := cmd.Flags().GetString("timestamp")
	if err != nil {
		return commonFlags{}, err
	}
	if timestamp == "" {
		timestamp = time.Now().Format("2006-01-02T15:04:05Z07:00")
	}
	timestmp, err := time.Parse("2006-01-02T15:04:05Z07:00", timestamp)
	if err != nil {
		return commonFlags{}, err
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return commonFlags{}, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return commonFlags{}, err
	}
	out, err := cmd.Flags().GetString("out")
	if err != nil {
		return commonFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return commonFlags{}, err
	}
	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	return commonFlags{
		rawImage:  rawImage,
		pki:       pki,
		variant:   variant,
		version:   ver,
		timestamp: timestmp,
		region:    region,
		bucket:    bucket,
		out:       out,
		logLevel:  logLevel,
	}, nil
}

type awsFlags struct {
	commonFlags
	awsRegion string
	awsBucket string
}

func parseAWSFlags(cmd *cobra.Command) (awsFlags, error) {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return awsFlags{}, err
	}

	awsRegion, err := cmd.Flags().GetString("aws-region")
	if err != nil {
		return awsFlags{}, err
	}
	awsBucket, err := cmd.Flags().GetString("aws-bucket")
	if err != nil {
		return awsFlags{}, err
	}

	common.provider = cloudprovider.AWS
	return awsFlags{
		commonFlags: common,
		awsRegion:   awsRegion,
		awsBucket:   awsBucket,
	}, nil
}

type azureFlags struct {
	commonFlags
	azSubscription  string
	azLocation      string
	azResourceGroup string
}

func parseAzureFlags(cmd *cobra.Command) (azureFlags, error) {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return azureFlags{}, err
	}

	azSubscription, err := cmd.Flags().GetString("az-subscription")
	if err != nil {
		return azureFlags{}, err
	}
	azLocation, err := cmd.Flags().GetString("az-location")
	if err != nil {
		return azureFlags{}, err
	}
	azResourceGroup, err := cmd.Flags().GetString("az-resource-group")
	if err != nil {
		return azureFlags{}, err
	}

	common.provider = cloudprovider.Azure
	return azureFlags{
		commonFlags:     common,
		azSubscription:  azSubscription,
		azLocation:      azLocation,
		azResourceGroup: azResourceGroup,
	}, nil
}

type gcpFlags struct {
	commonFlags
	gcpProject  string
	gcpLocation string
	gcpBucket   string
}

func parseGCPFlags(cmd *cobra.Command) (gcpFlags, error) {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return gcpFlags{}, err
	}

	gcpProject, err := cmd.Flags().GetString("gcp-project")
	if err != nil {
		return gcpFlags{}, err
	}
	gcpLocation, err := cmd.Flags().GetString("gcp-location")
	if err != nil {
		return gcpFlags{}, err
	}
	gcpBucket, err := cmd.Flags().GetString("gcp-bucket")
	if err != nil {
		return gcpFlags{}, err
	}

	common.provider = cloudprovider.GCP
	return gcpFlags{
		commonFlags: common,
		gcpProject:  gcpProject,
		gcpLocation: gcpLocation,
		gcpBucket:   gcpBucket,
	}, nil
}
