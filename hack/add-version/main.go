/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// add-version adds a new constellation release version to the list of available versions.
// It is meant to be run by the CI pipeline to make new versions available / discoverable.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/update"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/semver"
)

var errVersionListMissing = errors.New("version list does not exist")

const (
	stream                       = "stable"
	imageKind                    = "image"
	defaultRegion                = "eu-central-1"
	defaultBucket                = "cdn-constellation-backend"
	defaultDistributionID        = "E1H77EZTHC3NE4"
	maxCacheInvalidationWaitTime = 5 * time.Minute
)

func main() {
	version := flag.String("version", "", "Version to add (format: \"v1.2.3\")")
	region := flag.String("region", defaultRegion, "AWS region")
	bucket := flag.String("bucket", defaultBucket, "S3 bucket")
	distributionID := flag.String("distribution-id", defaultDistributionID, "cloudfront distribution id")
	flag.Parse()

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)
	if err := validateVersion(*version); err != nil {
		log.With(zap.Error(err)).Fatalf("Invalid version")
	}
	major := semver.Major(*version)
	minor := semver.MajorMinor(*version)

	ctx := context.Background()

	updateFetcher := update.New()
	versionManager, err := newVersionManager(ctx, *region, *bucket, *distributionID)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create version uploader")
	}

	// ensure minor version exists in list for base major version
	minorVersions, err := versionManager.getMinorVersions(ctx, *version)
	if err != nil {
		if !errors.Is(err, errVersionListMissing) {
			log.With(zap.Error(err)).Fatalf("Failed to get minor versions")
		}
		log.Infof("Version list for minor versions under %q does not exist. Creating new list.", major)
		minorVersions = &update.VersionsList{
			Stream:      stream,
			Granularity: "major",
			Base:        major,
			Kind:        imageKind,
			Versions:    []string{},
		}
	}
	if minorVersions.Contains(minor) {
		log.Infof("Version %q already exists in list %v.", minor, minorVersions.Versions)
	} else {
		if err := versionManager.addMinorVersion(ctx, *version, minorVersions); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to add minor version")
		}
		log.Infof("Added %q to list.", minor)
	}

	// ensure patch version exists in list for base minor version
	patchVersions, err := versionManager.getPatchVersions(ctx, *version)
	if err != nil {
		if !errors.Is(err, errVersionListMissing) {
			log.With(zap.Error(err)).Fatalf("Failed to get patch versions")
		}
		log.Infof("Version list for patch versions under %q does not exist. Creating new list.", minor)
		patchVersions = &update.VersionsList{
			Stream:      stream,
			Granularity: "minor",
			Base:        minor,
			Kind:        imageKind,
			Versions:    []string{},
		}
	}
	if patchVersions.Contains(*version) {
		log.Infof("Version %q already exists in list %v.", *version, patchVersions.Versions)
	} else {
		if err := versionManager.addPatchVersion(ctx, *version, patchVersions); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to add patch version")
		}
		log.Infof("Added %q to list.", *version)
	}

	log.Infof("Successfully added version %q at the following URLs:", *version)
	log.Infof("major to minor url: %s", versionURL("major", major))
	log.Infof("minor to patch url: %s", versionURL("minor", minor))

	log.Infof("Waiting for cache invalidation.")
	if err := versionManager.invalidateCaches(ctx, *version); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to invalidate caches")
	}

	sawAddedVersions := true
	if err := ensureMinorVersionExists(ctx, updateFetcher, *version); err != nil {
		sawAddedVersions = false
		log.Warnf("Failed to ensure minor version exists: %v. This may be resolved by waiting.", err)
	}

	if err := ensurePatchVersionExists(ctx, updateFetcher, *version); err != nil {
		sawAddedVersions = false
		log.Warnf("Failed to ensure patch version exists: %v. This may be resolved by waiting.", err)
	}

	if sawAddedVersions {
		log.Infof("Versions are available via API.")
	}
}

func validateVersion(version string) error {
	if !semver.IsValid(version) {
		return fmt.Errorf("version %q is not a valid semantic version", version)
	}
	if semver.Canonical(version) != version {
		return fmt.Errorf("version %q is not a canonical semantic version", version)
	}
	return nil
}

func ensureMinorVersionExists(ctx context.Context, fetcher *update.VersionsFetcher, version string) error {
	major := semver.Major(version)
	minor := semver.MajorMinor(version)
	existingMinorVersions, err := fetcher.MinorVersionsOf(ctx, stream, major, imageKind)
	if err != nil {
		return err
	}
	if !existingMinorVersions.Contains(minor) {
		return errors.New("minor version does not exist")
	}
	return nil
}

func ensurePatchVersionExists(ctx context.Context, fetcher *update.VersionsFetcher, version string) error {
	minor := semver.MajorMinor(version)
	patch := semver.Canonical(version)
	existingPatchVersions, err := fetcher.PatchVersionsOf(ctx, stream, minor, imageKind)
	if err != nil {
		return err
	}
	if !existingPatchVersions.Contains(patch) {
		return errors.New("patch version does not exist")
	}
	return nil
}

type versionManager struct {
	config         aws.Config
	cloudfrontc    *cloudfront.Client
	s3c            *s3.Client
	uploader       *s3manager.Uploader
	bucket         string
	distributionID string
}

func newVersionManager(ctx context.Context, region, bucket, distributionID string) (*versionManager, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	cloudfrontc := cloudfront.NewFromConfig(cfg)
	s3c := s3.NewFromConfig(cfg)
	uploader := s3manager.NewUploader(s3c)
	return &versionManager{
		config:         cfg,
		cloudfrontc:    cloudfrontc,
		s3c:            s3c,
		uploader:       uploader,
		bucket:         bucket,
		distributionID: distributionID,
	}, nil
}

func (m *versionManager) getMinorVersions(ctx context.Context, version string) (*update.VersionsList, error) {
	baseVersion := semver.Major(version)
	return m.getVersions(ctx, baseVersion)
}

func (m *versionManager) getPatchVersions(ctx context.Context, version string) (*update.VersionsList, error) {
	baseVersion := semver.MajorMinor(version)
	return m.getVersions(ctx, baseVersion)
}

func (m *versionManager) addMinorVersion(ctx context.Context, version string, minorVersions *update.VersionsList) error {
	baseVersion := semver.Major(version)
	minorVersion := semver.MajorMinor(version)
	return m.addVersion(ctx, baseVersion, minorVersion, minorVersions)
}

func (m *versionManager) addPatchVersion(ctx context.Context, version string, patchVersions *update.VersionsList) error {
	baseVersion := semver.MajorMinor(version)
	patchVersion := semver.Canonical(version)
	return m.addVersion(ctx, baseVersion, patchVersion, patchVersions)
}

func (m *versionManager) getVersions(ctx context.Context, baseVersion string) (*update.VersionsList, error) {
	granularity, err := granularityFromVersion(baseVersion)
	if err != nil {
		return nil, err
	}
	out, err := m.s3c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(versionJSONPath(granularity, baseVersion)),
	})
	if err != nil {
		var nosuchkey *s3types.NoSuchKey
		if errors.As(err, &nosuchkey) {
			return nil, errVersionListMissing
		}
		return nil, err
	}
	defer out.Body.Close()
	var versions update.VersionsList
	if err := json.NewDecoder(out.Body).Decode(&versions); err != nil {
		return nil, err
	}
	return &versions, nil
}

func (m *versionManager) addVersion(ctx context.Context, baseVersion, version string, list *update.VersionsList) error {
	granularity, err := granularityFromVersion(baseVersion)
	if err != nil {
		return err
	}
	list.Versions = append(list.Versions, version)
	sort.Strings(list.Versions)
	if err := list.Validate(); err != nil {
		return err
	}
	rawList, err := json.Marshal(list)
	if err != nil {
		return err
	}
	_, err = m.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(versionJSONPath(granularity, baseVersion)),
		Body:   bytes.NewBuffer(rawList),
	})
	return err
}

func (m *versionManager) invalidateCaches(ctx context.Context, version string) error {
	major := semver.Major(version)
	minor := semver.MajorMinor(version)
	invalidation, err := m.cloudfrontc.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(m.distributionID),
		InvalidationBatch: &cftypes.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			Paths: &cftypes.Paths{
				Quantity: aws.Int32(2),
				Items: []string{
					"/" + versionJSONPath("major", major),
					"/" + versionJSONPath("minor", minor),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	waiter := cloudfront.NewInvalidationCompletedWaiter(m.cloudfrontc)
	if err := waiter.Wait(ctx, &cloudfront.GetInvalidationInput{
		DistributionId: aws.String(m.distributionID),
		Id:             invalidation.Invalidation.Id,
	}, maxCacheInvalidationWaitTime); err != nil {
		return err
	}
	return nil
}

func granularityFromVersion(version string) (string, error) {
	switch {
	case semver.Major(version) == version:
		return "major", nil
	case semver.MajorMinor(version) == version:
		return "minor", nil
	case semver.Canonical(version) == version:
		return "patch", nil
	default:
		return "", fmt.Errorf("invalid version %q", version)
	}
}

func versionJSONPath(granularity, base string) string {
	return path.Join(constants.CDNUpdatesPath, stream, granularity, base, imageKind+".json")
}

func versionURL(granularity, base string) string {
	return constants.CDNRepositoryURL + "/" + versionJSONPath(granularity, base)
}
