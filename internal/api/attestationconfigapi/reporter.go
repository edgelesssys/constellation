/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
The reporter contains the logic to determine a latest version for Azure SEVSNP based on cached version values observed on CVM instances.
Some code in this file (e.g. listing cached files) does not rely on dedicated API objects and instead uses the AWS SDK directly,
for no other reason than original development speed.
*/
package attestationconfigapi

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
)

// cachedVersionsSubDir is the subdirectory in the bucket where the cached versions are stored.
const cachedVersionsSubDir = "cached-versions"

// ErrNoNewerVersion is returned if the input version is not newer than the latest API version.
var ErrNoNewerVersion = errors.New("input version is not newer than latest API version")

func reportVersionDir(attestation variant.Variant) string {
	return path.Join(AttestationURLPath, attestation.String(), cachedVersionsSubDir)
}

// UploadSEVSNPVersionLatest saves the given version to the cache, determines the smallest
// TCB version in the cache among the last cacheWindowSize versions and updates
// the latest version in the API if there is an update.
// force can be used to bypass the validation logic against the cached versions.
func (c Client) UploadSEVSNPVersionLatest(ctx context.Context, attestation variant.Variant, inputVersion,
	latestAPIVersion SEVSNPVersion, now time.Time, force bool,
) error {
	if err := c.cacheSEVSNPVersion(ctx, attestation, inputVersion, now); err != nil {
		return fmt.Errorf("reporting version: %w", err)
	}
	if force {
		return c.uploadSEVSNPVersion(ctx, attestation, inputVersion, now)
	}
	versionDates, err := c.listCachedVersions(ctx, attestation)
	if err != nil {
		return fmt.Errorf("list reported versions: %w", err)
	}
	if len(versionDates) < c.cacheWindowSize {
		c.s3Client.Logger.Warn(fmt.Sprintf("Skipping version update, found %d, expected %d reported versions.", len(versionDates), c.cacheWindowSize))
		return nil
	}
	minVersion, minDate, err := c.findMinVersion(ctx, attestation, versionDates)
	if err != nil {
		return fmt.Errorf("get minimal version: %w", err)
	}
	c.s3Client.Logger.Info(fmt.Sprintf("Found minimal version: %+v with date: %s", minVersion, minDate))
	shouldUpdateAPI, err := isInputNewerThanOtherVersion(minVersion, latestAPIVersion)
	if err != nil {
		return ErrNoNewerVersion
	}
	if !shouldUpdateAPI {
		c.s3Client.Logger.Info(fmt.Sprintf("Input version: %+v is not newer than latest API version: %+v", minVersion, latestAPIVersion))
		return nil
	}
	c.s3Client.Logger.Info(fmt.Sprintf("Input version: %+v is newer than latest API version: %+v", minVersion, latestAPIVersion))
	t, err := time.Parse(VersionFormat, minDate)
	if err != nil {
		return fmt.Errorf("parsing date: %w", err)
	}
	if err := c.uploadSEVSNPVersion(ctx, attestation, minVersion, t); err != nil {
		return fmt.Errorf("uploading version: %w", err)
	}
	c.s3Client.Logger.Info(fmt.Sprintf("Successfully uploaded new Azure SEV-SNP version: %+v", minVersion))
	return nil
}

// cacheSEVSNPVersion uploads the latest observed version numbers of the Azure SEVSNP. This version is used to later report the latest version numbers to the API.
func (c Client) cacheSEVSNPVersion(ctx context.Context, attestation variant.Variant, version SEVSNPVersion, date time.Time) error {
	dateStr := date.Format(VersionFormat) + ".json"
	res := putCmd{
		apiObject: reportedSEVSNPVersionAPI{Version: dateStr, variant: attestation, SEVSNPVersion: version},
		signer:    c.signer,
	}
	return res.Execute(ctx, c.s3Client)
}

func (c Client) listCachedVersions(ctx context.Context, attestation variant.Variant) ([]string, error) {
	list, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketID),
		Prefix: aws.String(reportVersionDir(attestation)),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}
	var dates []string
	for _, obj := range list.Contents {
		fileName := path.Base(*obj.Key)
		if strings.HasSuffix(fileName, ".json") {
			dates = append(dates, fileName[:len(fileName)-5])
		}
	}
	return dates, nil
}

// findMinVersion finds the minimal version of the given version dates among the latest values in the version window size.
func (c Client) findMinVersion(ctx context.Context, attesation variant.Variant, versionDates []string) (SEVSNPVersion, string, error) {
	var minimalVersion *SEVSNPVersion
	var minimalDate string
	sort.Sort(sort.Reverse(sort.StringSlice(versionDates))) // sort in reverse order to slice the latest versions
	versionDates = versionDates[:c.cacheWindowSize]
	sort.Strings(versionDates) // sort with oldest first to to take the minimal version with the oldest date
	for _, date := range versionDates {
		obj, err := client.Fetch(ctx, c.s3Client, reportedSEVSNPVersionAPI{Version: date + ".json", variant: attesation})
		if err != nil {
			return SEVSNPVersion{}, "", fmt.Errorf("get object: %w", err)
		}
		// Need to set this explicitly as the variant is not part of the marshalled JSON.
		obj.variant = attesation

		if minimalVersion == nil {
			minimalVersion = &obj.SEVSNPVersion
			minimalDate = date
		} else {
			shouldUpdateMinimal, err := isInputNewerThanOtherVersion(*minimalVersion, obj.SEVSNPVersion)
			if err != nil {
				continue
			}
			if shouldUpdateMinimal {
				minimalVersion = &obj.SEVSNPVersion
				minimalDate = date
			}
		}
	}
	return *minimalVersion, minimalDate, nil
}

// isInputNewerThanOtherVersion compares all version fields and returns true if any input field is newer.
func isInputNewerThanOtherVersion(input, other SEVSNPVersion) (bool, error) {
	if input == other {
		return false, nil
	}
	if input.TEE < other.TEE {
		return false, fmt.Errorf("input TEE version: %d is older than latest API version: %d", input.TEE, other.TEE)
	}
	if input.SNP < other.SNP {
		return false, fmt.Errorf("input SNP version: %d is older than latest API version: %d", input.SNP, other.SNP)
	}
	if input.Microcode < other.Microcode {
		return false, fmt.Errorf("input Microcode version: %d is older than latest API version: %d", input.Microcode, other.Microcode)
	}
	if input.Bootloader < other.Bootloader {
		return false, fmt.Errorf("input Bootloader version: %d is older than latest API version: %d", input.Bootloader, other.Bootloader)
	}
	return true, nil
}

// reportedSEVSNPVersionAPI is the request to get the version information of the specific version in the config api.
type reportedSEVSNPVersionAPI struct {
	Version string          `json:"-"`
	variant variant.Variant `json:"-"`
	SEVSNPVersion
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i reportedSEVSNPVersionAPI) JSONPath() string {
	return path.Join(reportVersionDir(i.variant), i.Version)
}

// ValidateRequest validates the request.
func (i reportedSEVSNPVersionAPI) ValidateRequest() error {
	if !strings.HasSuffix(i.Version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op at the moment.
func (i reportedSEVSNPVersionAPI) Validate() error {
	return nil
}
