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
	"github.com/edgelesssys/constellation/v2/internal/versionsapi-old"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/semver"
)

var errVersionListMissing = errors.New("version list does not exist")

const (
	skipRefStr                   = "-"
	imageKind                    = "image"
	defaultRegion                = "eu-central-1"
	defaultBucket                = "cdn-constellation-backend"
	defaultDistributionID        = "E1H77EZTHC3NE4"
	maxCacheInvalidationWaitTime = 5 * time.Minute
)

func main() {
	log := logger.New(logger.JSONLog, zapcore.InfoLevel)
	ctx := context.Background()

	flags := flags{
		version:        flag.String("version", "", "Version to add (format: \"v1.2.3\")"),
		stream:         flag.String("stream", "", "Stream to add the version to"),
		ref:            flag.String("ref", "", "Ref to add the version to"),
		release:        flag.Bool("release", false, "Whether the version is a release"),
		latest:         flag.Bool("latest", false, "Whether to set this version as the new latest version"),
		dryRun:         flag.Bool("dryrun", false, "Whether to run in dry-run mode (no changes are made)"),
		region:         flag.String("region", defaultRegion, "AWS region"),
		bucket:         flag.String("bucket", defaultBucket, "S3 bucket"),
		distributionID: flag.String("distribution-id", defaultDistributionID, "cloudfront distribution id"),
	}
	flag.Parse()
	if err := flags.validate(); err != nil {
		log.With(zap.Error(err)).Fatalf("Invalid flags")
	}

	updateFetcher := versionsapi.New()
	versionManager, err := newVersionManager(ctx, *flags.region, *flags.bucket, *flags.distributionID, *flags.dryRun, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create version uploader")
	}

	ver := version{
		versionStr: *flags.version,
		stream:     *flags.stream,
		ref:        *flags.ref,
	}

	if err := ensureMinorVersion(ctx, versionManager, ver, log); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to ensure minor version")
	}

	added, err := ensurePatchVersion(ctx, versionManager, ver, log)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to ensure patch version")
	}

	if added && *flags.latest {
		if err := versionManager.addLatest(ctx, ver); err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to update latest version object")
		}
		log.Infof("Added %q as latest version.", ver)
	}

	log.Infof("Major to minor url: %s", ver.URL(granularityMajor))
	log.Infof("Minor to patch url: %s", ver.URL(granularityMinor))

	if !versionManager.dirty {
		log.Infof("No changes made, everything up to date.")
		return
	}
	log.Infof("Successfully added version %q", *flags.version)

	log.Infof("Waiting for cache invalidation.")
	if err := versionManager.invalidateCaches(ctx, ver, *flags.latest); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to invalidate caches")
	}

	waitForCacheUpdate(ctx, updateFetcher, ver, log)
}

func ensureMinorVersion(ctx context.Context, versionManager *versionManager, ver version, log *logger.Logger) error {
	minorVerList, err := versionManager.getVersionList(ctx, ver, granularityMajor)
	log.Debugf("Minor version list: %v", minorVerList)
	if errors.Is(err, errVersionListMissing) {
		log.Infof("Version list for minor versions under %q does not exist. Creating new list.", ver.Major())
		minorVerList = &versionsapi.List{
			Ref:         ver.Ref(),
			Stream:      ver.Stream(),
			Granularity: "major",
			Base:        ver.Major(),
			Kind:        imageKind,
			Versions:    []string{},
		}
	} else if err != nil {
		return fmt.Errorf("failed to list minor versions: %w", err)
	}

	if minorVerList.Contains(ver.MajorMinor()) {
		log.Infof("Version %q already exists in list %v.", ver.MajorMinor(), minorVerList.Versions)
		return nil
	}

	minorVerList.Versions = append(minorVerList.Versions, ver.MajorMinor())
	log.Debugf("New minor version list: %v", minorVerList)

	if err := versionManager.updateVersionList(ctx, minorVerList); err != nil {
		return fmt.Errorf("failed to add minor version: %w", err)
	}

	log.Infof("Added %q to list.", ver.MajorMinor())
	return nil
}

func ensurePatchVersion(ctx context.Context, versionManager *versionManager, ver version, log *logger.Logger) (bool, error) {
	pathVerList, err := versionManager.getVersionList(ctx, ver, granularityMinor)
	if errors.Is(err, errVersionListMissing) {
		log.Infof("Version list for patch versions under %q does not exist. Creating new list.", ver.MajorMinor())
		pathVerList = &versionsapi.List{
			Ref:         ver.Ref(),
			Stream:      ver.Stream(),
			Granularity: "minor",
			Base:        ver.MajorMinor(),
			Kind:        imageKind,
			Versions:    []string{},
		}
	} else if err != nil {
		return false, fmt.Errorf("failed to get patch versions: %w", err)
	}

	if pathVerList.Contains(ver.String()) {
		log.Infof("Version %q already exists in list %v.", ver.String(), pathVerList.Versions)
		return false, nil
	}

	pathVerList.Versions = append(pathVerList.Versions, ver.String())

	if err := versionManager.updateVersionList(ctx, pathVerList); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to add patch version")
	}

	log.Infof("Added %q to list.", ver.String())
	return true, nil
}

type version struct {
	versionStr string
	stream     string
	ref        string
}

func (v *version) String() string {
	return semver.Canonical(v.versionStr)
}

func (v *version) Major() string {
	return semver.Major(v.versionStr)
}

func (v *version) MajorMinor() string {
	return semver.MajorMinor(v.versionStr)
}

func (v *version) WithGranularity(gran granularity) string {
	switch gran {
	case granularityMajor:
		return v.Major()
	case granularityMinor:
		return v.MajorMinor()
	default:
		return ""
	}
}

func (v *version) URL(gran granularity) string {
	return constants.CDNRepositoryURL + "/" + v.JSONPath(gran)
}

func (v *version) JSONPath(gran granularity) string {
	return path.Join(constants.CDNAPIPrefix, "ref", v.ref, "stream", v.stream, "versions", gran.String(), v.WithGranularity(gran), imageKind+".json")
}

func (v *version) Stream() string {
	return v.stream
}

func (v *version) Ref() string {
	return v.ref
}

type flags struct {
	version        *string
	stream         *string
	ref            *string
	release        *bool
	latest         *bool
	dryRun         *bool
	region         *string
	bucket         *string
	distributionID *string
}

func (f *flags) validate() error {
	if err := validateVersion(*f.version); err != nil {
		return err
	}

	if *f.ref == "" && !*f.release {
		if !*f.release {
			return fmt.Errorf("branch flag must be set for non-release versions")
		}
	}

	if *f.ref != "" && *f.release {
		return fmt.Errorf("branch flag must not be set for release versions")
	}

	if *f.release {
		*f.ref = skipRefStr
	} else {
		*f.latest = true // always set latest for non-release versions
	}

	ref := versionsapi.CanonicalRef(*f.ref)
	if !versionsapi.IsValidRef(ref) {
		return fmt.Errorf("invalid ref %q", *f.ref)
	}
	*f.ref = ref

	if !versionsapi.IsValidStream(*f.ref, *f.stream) {
		return fmt.Errorf("invalid stream %q for ref %q", *f.stream, *f.ref)
	}

	return nil
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

func ensureMinorVersionExists(ctx context.Context, fetcher *versionsapi.Fetcher, ver version) error {
	existingMinorVersions, err := fetcher.MinorVersionsOf(ctx, ver.Ref(), ver.Stream(), ver.Major(), imageKind)
	if err != nil {
		return err
	}
	if !existingMinorVersions.Contains(ver.MajorMinor()) {
		return errors.New("minor version does not exist")
	}
	return nil
}

func ensurePatchVersionExists(ctx context.Context, fetcher *versionsapi.Fetcher, ver version) error {
	existingPatchVersions, err := fetcher.PatchVersionsOf(ctx, ver.Ref(), ver.Stream(), ver.MajorMinor(), imageKind)
	if err != nil {
		return err
	}
	if !existingPatchVersions.Contains(ver.String()) {
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
	dirty          bool // manager gets dirty on write
	dryRun         bool // no write operations
	log            *logger.Logger
}

func newVersionManager(ctx context.Context, region, bucket, distributionID string, dryRun bool, log *logger.Logger) (*versionManager, error) {
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
		dryRun:         dryRun,
		log:            log,
	}, nil
}

func (m *versionManager) getVersionList(ctx context.Context, ver version, gran granularity) (*versionsapi.List, error) {
	in := &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(ver.JSONPath(gran)),
	}
	out, err := m.s3c.GetObject(ctx, in)
	var noSuchkey *s3types.NoSuchKey
	if errors.As(err, &noSuchkey) {
		return nil, errVersionListMissing
	} else if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	var list versionsapi.List
	if err := json.NewDecoder(out.Body).Decode(&list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (m *versionManager) updateVersionList(ctx context.Context, list *versionsapi.List) error {
	semver.Sort(list.Versions)
	if err := list.Validate(); err != nil {
		return err
	}

	rawList, err := json.Marshal(list)
	if err != nil {
		return err
	}

	in := &s3.PutObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(list.JSONPath()),
		Body:   bytes.NewBuffer(rawList),
	}

	if m.dryRun {
		m.log.Infof("dryRun: s3 put object {Bucket: %v, Key: %v, Body: %v", m.bucket, list.JSONPath(), string(rawList))
		return nil
	}

	m.dirty = true

	_, err = m.uploader.Upload(ctx, in)

	return err
}

func (m *versionManager) addLatest(ctx context.Context, ver version) error {
	latest := &versionsapi.Latest{
		Ref:     ver.Ref(),
		Stream:  ver.Stream(),
		Kind:    imageKind,
		Version: ver.String(),
	}
	if err := latest.Validate(); err != nil {
		return err
	}

	rawLatest, err := json.Marshal(latest)
	if err != nil {
		return err
	}

	in := &s3.PutObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(latest.JSONPath()),
		Body:   bytes.NewBuffer(rawLatest),
	}

	if m.dryRun {
		m.log.Infof("dryRun: s3 put object {Bucket: %v, Key: %v, Body: %v", m.bucket, latest.JSONPath(), string(rawLatest))
		return nil
	}

	m.dirty = true

	_, err = m.uploader.Upload(ctx, in)

	return err
}

func (m *versionManager) invalidateCaches(ctx context.Context, ver version, latest bool) error {
	invalidIn := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(m.distributionID),
		InvalidationBatch: &cftypes.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			Paths: &cftypes.Paths{
				Quantity: aws.Int32(2),
				Items: []string{
					"/" + ver.URL(granularityMajor),
					"/" + ver.URL(granularityMinor),
				},
			},
		},
	}
	if latest {
		invalidIn.InvalidationBatch.Paths.Quantity = aws.Int32(3)
		path := path.Join("ref", ver.Ref(), "stream", ver.Stream(), "versions/latest/image.json")
		invalidIn.InvalidationBatch.Paths.Items = append(invalidIn.InvalidationBatch.Paths.Items, "/"+path)
	}
	invalidation, err := m.cloudfrontc.CreateInvalidation(ctx, invalidIn)
	if err != nil {
		return err
	}

	waiter := cloudfront.NewInvalidationCompletedWaiter(m.cloudfrontc)
	waitIn := &cloudfront.GetInvalidationInput{
		DistributionId: aws.String(m.distributionID),
		Id:             invalidation.Invalidation.Id,
	}
	if err := waiter.Wait(ctx, waitIn, maxCacheInvalidationWaitTime); err != nil {
		return err
	}

	return nil
}

func waitForCacheUpdate(ctx context.Context, updateFetcher *versionsapi.Fetcher, ver version, log *logger.Logger) {
	sawAddedVersions := true
	if err := ensureMinorVersionExists(ctx, updateFetcher, ver); err != nil {
		sawAddedVersions = false
		log.Warnf("Failed to ensure minor version exists: %v. This may be resolved by waiting.", err)
	}

	if err := ensurePatchVersionExists(ctx, updateFetcher, ver); err != nil {
		sawAddedVersions = false
		log.Warnf("Failed to ensure patch version exists: %v. This may be resolved by waiting.", err)
	}

	if sawAddedVersions {
		log.Infof("Versions are available via API.")
	}
}

type granularity int

const (
	granularityMajor granularity = iota
	granularityMinor
)

func (g granularity) String() string {
	switch g {
	case granularityMajor:
		return "major"
	case granularityMinor:
		return "minor"
	default:
		return "unknown"
	}
}
