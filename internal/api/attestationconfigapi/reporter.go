/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
The reporter contains the logic to determine a latest version for Azure SEVSNP based on cached version values observed on CVM instances.
*/
package attestationconfigapi

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/edgelesssys/constellation/v2/internal/api/client"
	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
)

// cachedVersionsSubDir is the subdirectory in the bucket where the cached versions are stored.
// TODO(elchead): store in a different directory so that it is not mirrored to the CDN?
const cachedVersionsSubDir = "cached-versions"

// versionWindowSize defines the number of versions to be considered for the latest version. Each week 5 versions are uploaded for each node of the verify cluster.
const versionWindowSize = 15

// timeFrameForCachedVersions defines the time frame for reported versions which are considered to define the latest version.
const timeFrameForCachedVersions = 21 * 24 * time.Hour

var reportVersionDir = path.Join(attestationURLPath, variant.AzureSEVSNP{}.String(), cachedVersionsSubDir)

// UpdateLatestVersion reports the given version, checks the reported version values
// and updates the latest version of the Azure SEVSNP in the API if there is an update .
// force can be used to force an update of the latest API version regardless of the cache reporting.
func (c Client) UpdateLatestVersion(ctx context.Context, inputVersion,
	latestAPIVersion AzureSEVSNPVersion, now time.Time, force bool,
) error {
	if err := c.reportAzureSEVSNPVersion(ctx, inputVersion, now); err != nil {
		return fmt.Errorf("reporting version: %w", err)
	}
	if force {
		if err := c.uploadAzureSEVSNPVersion(ctx, inputVersion, now); err != nil {
			return fmt.Errorf("force uploading version: %w", err)
		}
		return nil
	}
	versionDates, err := c.listReportedVersions(ctx, timeFrameForCachedVersions, now)
	if err != nil {
		return fmt.Errorf("list reported versions: %w", err)
	}
	if len(versionDates) < versionWindowSize {
		c.s3Client.Logger.Infof("Skipping version update since found only %d out of expected reported versions.", len(versionDates), versionWindowSize)
		return nil
	}
	minVersion, minDate, err := c.findMinVersion(ctx, versionDates)
	if err != nil {
		return fmt.Errorf("get minimal version: %w", err)
	}
	c.s3Client.Logger.Infof("Found minimal version: %+v with date: %s", *minVersion, minDate)
	shouldUpdateAPI, err := isInputNewerThanOtherVersion(*minVersion, latestAPIVersion)
	if err == nil && shouldUpdateAPI {
		c.s3Client.Logger.Infof("Input version: %+v is newer than latest API version: %+v", *minVersion, latestAPIVersion)
		t, err := time.Parse(VersionFormat, minDate)
		if err != nil {
			return fmt.Errorf("parsing date: %w", err)
		}
		if err := c.uploadAzureSEVSNPVersion(ctx, *minVersion, t); err != nil {
			return fmt.Errorf("uploading version: %w", err)
		}
		return nil
	}
	c.s3Client.Logger.Infof("Input version: %+v is not newer than latest API version: %+v", *minVersion, latestAPIVersion)
	return nil
}

// reportAzureSEVSNPVersion uploads the latest observed version numbers of the Azure SEVSNP. This version is used to later report the latest version numbers to the API.
func (c Client) reportAzureSEVSNPVersion(ctx context.Context, version AzureSEVSNPVersion, date time.Time) error {
	dateStr := date.Format(VersionFormat) + ".json"
	res := putCmdWithoutSigning{
		apiObject: reportedAzureSEVSNPVersionAPI{Version: dateStr, AzureSEVSNPVersion: version},
	}
	return res.Execute(ctx, c.s3Client)
}

func (c Client) listReportedVersions(ctx context.Context, _ time.Duration, _ time.Time) ([]string, error) {
	list, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketID),
		Prefix: aws.String(reportVersionDir),
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
func (c Client) findMinVersion(ctx context.Context, versionDates []string) (*AzureSEVSNPVersion, string, error) {
	var minimalVersion *AzureSEVSNPVersion
	var minimalDate string
	sort.Strings(versionDates) // the oldest date with the minimal version should be taken
	versionDates = versionDates[:versionWindowSize]
	for _, date := range versionDates {
		obj, err := client.Fetch(ctx, c.s3Client, reportedAzureSEVSNPVersionAPI{Version: date + ".json"})
		if err != nil {
			return nil, "", fmt.Errorf("get object: %w", err)
		}

		if minimalVersion == nil {
			minimalVersion = &obj.AzureSEVSNPVersion
			minimalDate = date
		} else {
			shouldUpdateMinimal, err := isInputNewerThanOtherVersion(*minimalVersion, obj.AzureSEVSNPVersion)
			if err != nil {
				continue
			}
			if shouldUpdateMinimal {
				minimalVersion = &obj.AzureSEVSNPVersion
				minimalDate = date
			}
		}
	}
	return minimalVersion, minimalDate, nil
}

func filterDatesWithinTime(dates []string, now time.Time, timeFrame time.Duration) []string {
	var datesWithinTimeFrame []string
	for _, date := range dates {
		t, err := time.Parse(VersionFormat, date)
		if err != nil {
			continue
		}
		fmt.Println(now, " t ", t, " sub ", now.Sub(t))
		if now.Sub(t) >= 0 && now.Sub(t) <= timeFrame {
			datesWithinTimeFrame = append(datesWithinTimeFrame, date)
		}
	}
	return datesWithinTimeFrame
}

// isInputNewerThanOtherVersion compares all version fields and returns true if any input field is newer.
func isInputNewerThanOtherVersion(input, other AzureSEVSNPVersion) (bool, error) {
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

// reportedAzureSEVSNPVersionAPI is the request to get the version information of the specific version in the config api.
type reportedAzureSEVSNPVersionAPI struct {
	Version string `json:"-"`
	AzureSEVSNPVersion
}

// URL returns the URL for the request to the config api.
func (i reportedAzureSEVSNPVersionAPI) URL() (string, error) {
	return getURL(i)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i reportedAzureSEVSNPVersionAPI) JSONPath() string {
	return path.Join(reportVersionDir, i.Version)
}

// ValidateRequest validates the request.
func (i reportedAzureSEVSNPVersionAPI) ValidateRequest() error {
	if !strings.HasSuffix(i.Version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op at the moment.
func (i reportedAzureSEVSNPVersionAPI) Validate() error {
	return nil
}

type putCmdWithoutSigning struct {
	apiObject apiclient.APIObject
}

func (p putCmdWithoutSigning) Execute(ctx context.Context, c *apiclient.Client) error {
	return apiclient.Update(ctx, c, p.apiObject)
}
