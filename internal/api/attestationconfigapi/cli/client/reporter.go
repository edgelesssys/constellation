/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
)

// cachedVersionsSubDir is the subdirectory in the bucket where the cached versions are stored.
const cachedVersionsSubDir = "cached-versions"

// ErrNoNewerVersion is returned if the input version is not newer than the latest API version.
var ErrNoNewerVersion = errors.New("input version is not newer than latest API version")

func reportVersionDir(attestation variant.Variant) string {
	return path.Join(attestationconfigapi.AttestationURLPath, attestation.String(), cachedVersionsSubDir)
}

// UploadLatestVersion saves the given version to the cache, determines the smallest
// TCB version in the cache among the last cacheWindowSize versions and updates
// the latest version in the API if there is an update.
// force can be used to bypass the validation logic against the cached versions.
func (c Client) UploadLatestVersion(
	ctx context.Context, attestationVariant variant.Variant,
	inputVersion, latestVersionInAPI any,
	now time.Time, force bool,
) error {
	// Validate input versions against configured attestation variant
	// This allows us to skip these checks in the individual variant implementations
	var err error
	actionForVariant(attestationVariant,
		func() {
			if _, ok := inputVersion.(attestationconfigapi.TDXVersion); !ok {
				err = fmt.Errorf("input version %q is not a TDX version", inputVersion)
			}
			if _, ok := latestVersionInAPI.(attestationconfigapi.TDXVersion); !ok {
				err = fmt.Errorf("latest API version %q is not a TDX version", latestVersionInAPI)
			}
		},
		func() {
			if _, ok := inputVersion.(attestationconfigapi.SEVSNPVersion); !ok {
				err = fmt.Errorf("input version %q is not a SNP version", inputVersion)
			}
			if _, ok := latestVersionInAPI.(attestationconfigapi.SEVSNPVersion); !ok {
				err = fmt.Errorf("latest API version %q is not a SNP version", latestVersionInAPI)
			}
		},
	)
	if err != nil {
		return err
	}

	if err := c.addVersionToCache(ctx, attestationVariant, inputVersion, now); err != nil {
		return fmt.Errorf("adding version to cache: %w", err)
	}

	// If force is set, immediately update the latest version to the new version in the API.
	if force {
		return c.uploadAsLatestVersion(ctx, attestationVariant, inputVersion, now)
	}

	// Otherwise, check the cached versions and update the latest version in the API if necessary.
	versionDates, err := c.listCachedVersions(ctx, attestationVariant)
	if err != nil {
		return fmt.Errorf("listing existing cached versions: %w", err)
	}
	if len(versionDates) < c.cacheWindowSize {
		c.log.Warn(fmt.Sprintf("Skipping version update, found %d, expected %d reported versions.", len(versionDates), c.cacheWindowSize))
		return nil
	}

	minVersion, minDate, err := c.findMinVersion(ctx, attestationVariant, versionDates)
	if err != nil {
		return fmt.Errorf("determining minimal version in cache: %w", err)
	}
	c.log.Info(fmt.Sprintf("Found minimal version: %+v with date: %s", minVersion, minDate))

	if !isInputNewerThanOtherVersion(attestationVariant, minVersion, latestVersionInAPI) {
		c.log.Info(fmt.Sprintf("Input version: %+v is not newer than latest API version: %+v. Skipping list update", minVersion, latestVersionInAPI))
		return ErrNoNewerVersion
	}

	c.log.Info(fmt.Sprintf("Input version: %+v is newer than latest API version: %+v", minVersion, latestVersionInAPI))
	t, err := time.Parse(VersionFormat, minDate)
	if err != nil {
		return fmt.Errorf("parsing date: %w", err)
	}

	if err := c.uploadAsLatestVersion(ctx, attestationVariant, minVersion, t); err != nil {
		return fmt.Errorf("uploading as latest version: %w", err)
	}

	c.log.Info(fmt.Sprintf("Successfully uploaded new %s version: %+v", attestationVariant, minVersion))
	return nil
}

// uploadAsLatestVersion uploads the given version and updates the list to set it as the "latest" version.
// The version's name is the UTC timestamp of the date.
// The /list entry stores the version name + .json suffix.
func (c Client) uploadAsLatestVersion(ctx context.Context, variant variant.Variant, inputVersion any, date time.Time) error {
	versions, err := c.List(ctx, variant)
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	if !variant.Equal(versions.Variant) {
		return nil
	}

	dateStr := date.Format(VersionFormat) + ".json"
	var ops []crudCmd

	obj := apiVersionObject{version: dateStr, variant: variant, cached: false}
	obj.setVersion(inputVersion)
	ops = append(ops, putCmd{
		apiObject: obj,
		signer:    c.signer,
	})

	versions.AddVersion(dateStr)

	ops = append(ops, putCmd{
		apiObject: versions,
		signer:    c.signer,
	})

	return executeAllCmds(ctx, c.s3Client, ops)
}

// addVersionToCache adds the given version to the cache.
func (c Client) addVersionToCache(ctx context.Context, variant variant.Variant, inputVersion any, date time.Time) error {
	dateStr := date.Format(VersionFormat) + ".json"
	obj := apiVersionObject{version: dateStr, variant: variant, cached: true}
	obj.setVersion(inputVersion)
	cmd := putCmd{
		apiObject: obj,
		signer:    c.signer,
	}
	return cmd.Execute(ctx, c.s3Client)
}

// findMinVersion returns the minimal version in the cache among the last cacheWindowSize versions.
func (c Client) findMinVersion(
	ctx context.Context, attestationVariant variant.Variant, versionDates []string,
) (any, string, error) {
	var getMinimalVersion func() (any, string, error)

	actionForVariant(attestationVariant,
		func() {
			getMinimalVersion = func() (any, string, error) {
				return findMinimalVersion[attestationconfigapi.TDXVersion](ctx, attestationVariant, versionDates, c.s3Client, c.cacheWindowSize)
			}
		},
		func() {
			getMinimalVersion = func() (any, string, error) {
				return findMinimalVersion[attestationconfigapi.SEVSNPVersion](ctx, attestationVariant, versionDates, c.s3Client, c.cacheWindowSize)
			}
		},
	)
	return getMinimalVersion()
}

func findMinimalVersion[T attestationconfigapi.TDXVersion | attestationconfigapi.SEVSNPVersion](
	ctx context.Context, variant variant.Variant, versionDates []string,
	s3Client *client.Client, cacheWindowSize int,
) (T, string, error) {
	var minimalVersion *T
	var minimalDate string
	sort.Sort(sort.Reverse(sort.StringSlice(versionDates))) // sort in reverse order to slice the latest versions
	versionDates = versionDates[:cacheWindowSize]
	sort.Strings(versionDates) // sort with oldest first to to take the minimal version with the oldest date

	for _, date := range versionDates {
		obj, err := client.Fetch(ctx, s3Client, apiVersionObject{version: date + ".json", variant: variant, cached: true})
		if err != nil {
			return *new(T), "", fmt.Errorf("get object: %w", err)
		}
		obj.variant = variant // variant is not set by Fetch, set it manually

		if minimalVersion == nil {
			v := obj.getVersion().(T)
			minimalVersion = &v
			minimalDate = date
			continue
		}

		// If the current minimal version has newer versions than the one we just fetched,
		// update the minimal version to the older version.
		if isInputNewerThanOtherVersion(variant, *minimalVersion, obj.getVersion()) {
			v := obj.getVersion().(T)
			minimalVersion = &v
			minimalDate = date
		}
	}

	return *minimalVersion, minimalDate, nil
}

func isInputNewerThanOtherVersion(variant variant.Variant, inputVersion, otherVersion any) bool {
	var result bool
	actionForVariant(variant,
		func() {
			input := inputVersion.(attestationconfigapi.TDXVersion)
			other := otherVersion.(attestationconfigapi.TDXVersion)
			result = isInputNewerThanOtherTDXVersion(input, other)
		},
		func() {
			input := inputVersion.(attestationconfigapi.SEVSNPVersion)
			other := otherVersion.(attestationconfigapi.SEVSNPVersion)
			result = isInputNewerThanOtherSEVSNPVersion(input, other)
		},
	)
	return result
}

type apiVersionObject struct {
	version string          `json:"-"`
	variant variant.Variant `json:"-"`
	cached  bool            `json:"-"`
	snp     attestationconfigapi.SEVSNPVersion
	tdx     attestationconfigapi.TDXVersion
}

func (a apiVersionObject) MarshalJSON() ([]byte, error) {
	var res []byte
	var err error
	actionForVariant(a.variant,
		func() {
			res, err = json.Marshal(a.tdx)
		},
		func() {
			res, err = json.Marshal(a.snp)
		},
	)
	return res, err
}

func (a *apiVersionObject) UnmarshalJSON(data []byte) error {
	errTDX := json.Unmarshal(data, &a.tdx)
	errSNP := json.Unmarshal(data, &a.snp)
	if errTDX == nil || errSNP == nil {
		return nil
	}
	return fmt.Errorf("trying to unmarshal data into both TDX and SNP versions: %w", errors.Join(errTDX, errSNP))
}

// JSONPath returns the path to the JSON file for the request to the config api.
// This is the path to the cached version in the S3 bucket.
func (a apiVersionObject) JSONPath() string {
	if a.cached {
		return path.Join(reportVersionDir(a.variant), a.version)
	}
	return path.Join(attestationconfigapi.AttestationURLPath, a.variant.String(), a.version)
}

// ValidateRequest validates the request.
func (a apiVersionObject) ValidateRequest() error {
	if !strings.HasSuffix(a.version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op.
func (a apiVersionObject) Validate() error {
	return nil
}

// getVersion returns the version.
func (a apiVersionObject) getVersion() any {
	var res any
	actionForVariant(a.variant,
		func() {
			res = a.tdx
		},
		func() {
			res = a.snp
		},
	)
	return res
}

// setVersion sets the version.
func (a *apiVersionObject) setVersion(version any) {
	actionForVariant(a.variant,
		func() {
			a.tdx = version.(attestationconfigapi.TDXVersion)
		},
		func() {
			a.snp = version.(attestationconfigapi.SEVSNPVersion)
		},
	)
}

// actionForVariant performs the given action based on the whether variant is a TDX or SEV-SNP variant.
func actionForVariant(
	attestationVariant variant.Variant,
	tdxAction func(), snpAction func(),
) {
	switch attestationVariant {
	case variant.AWSSEVSNP{}, variant.AzureSEVSNP{}, variant.GCPSEVSNP{}:
		snpAction()
	case variant.AzureTDX{}:
		tdxAction()
	default:
		panic(fmt.Sprintf("unsupported attestation variant: %s", attestationVariant))
	}
}

// isInputNewerThanOtherSEVSNPVersion compares all version fields and returns false if any input field is older, or the versions are equal.
func isInputNewerThanOtherSEVSNPVersion(input, other attestationconfigapi.SEVSNPVersion) bool {
	if input == other {
		return false
	}
	if input.TEE < other.TEE {
		return false
	}
	if input.SNP < other.SNP {
		return false
	}
	if input.Microcode < other.Microcode {
		return false
	}
	if input.Bootloader < other.Bootloader {
		return false
	}
	return true
}

// isInputNewerThanOtherSEVSNPVersion compares all version fields and returns false if any input field is older, or the versions are equal.
func isInputNewerThanOtherTDXVersion(input, other attestationconfigapi.TDXVersion) bool {
	if input == other {
		return false
	}

	if input.PCESVN < other.PCESVN {
		return false
	}
	if input.QESVN < other.QESVN {
		return false
	}

	// Validate component-wise security version numbers
	for idx, inputVersion := range input.TEETCBSVN {
		if inputVersion < other.TEETCBSVN[idx] {
			return false
		}
	}

	return true
}
