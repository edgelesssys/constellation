/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package imagefetcher provides helping wrappers around a versionsapi fetcher.

It also enables local image overrides and download of raw images.
*/
package imagefetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"regexp"

	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/mpimage"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/afero"
)

// Fetcher fetches image references using a lookup table.
type Fetcher struct {
	fetcher versionsAPIImageInfoFetcher
	fs      *afero.Afero
}

// New returns a new image fetcher.
func New() *Fetcher {
	return &Fetcher{
		fetcher: versionsapi.NewFetcher(),
		fs:      &afero.Afero{Fs: afero.NewOsFs()},
	}
}

// FetchReference fetches the image reference for a given image version uid, CSP and image variant.
func (f *Fetcher) FetchReference(ctx context.Context,
	provider cloudprovider.Provider, attestationVariant variant.Variant,
	image, region string, useMarketplaceImage bool,
) (string, error) {
	ver, err := versionsapi.NewVersionFromShortPath(image, versionsapi.VersionKindImage)
	if err != nil {
		return "", fmt.Errorf("parsing config image short path: %w", err)
	}

	if useMarketplaceImage {
		return buildMarketplaceImage(ver, provider)
	}

	imgInfoReq := versionsapi.ImageInfo{
		Ref:     ver.Ref(),
		Stream:  ver.Stream(),
		Version: ver.Version(),
	}

	url, err := imgInfoReq.URL()
	if err != nil {
		return "", err
	}

	imgInfo, err := getFromFile(f.fs, imgInfoReq)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		imgInfo, err = f.fetcher.FetchImageInfo(ctx, imgInfoReq)
	}

	var notFoundErr *fetcher.NotFoundError
	switch {
	case errors.As(err, &notFoundErr):
		overridePath := imageInfoFilename(imgInfoReq)
		return "", fmt.Errorf("image info file not found locally at %q or remotely at %s", overridePath, url)
	case err != nil:
		return "", err
	}

	if err := imgInfo.Validate(); err != nil {
		return "", fmt.Errorf("validating image info file: %w", err)
	}

	return getReferenceFromImageInfo(provider, attestationVariant.String(), imgInfo, filters(provider, region)...)
}

// buildMarketplaceImage returns a marketplace image URI for the given CSP and version.
func buildMarketplaceImage(ver versionsapi.Version, provider cloudprovider.Provider) (string, error) {
	sv, err := semver.New(ver.Version())
	if err != nil {
		return "", fmt.Errorf("parsing image version: %w", err)
	}

	switch provider {
	case cloudprovider.Azure:
		return mpimage.NewAzureMarketplaceImage(sv).URI(), nil
	default:
		return "", fmt.Errorf("marketplace images are not supported for csp %s", provider.String())
	}
}

func filters(provider cloudprovider.Provider, region string) []filter {
	var filters []filter
	switch provider {
	case cloudprovider.AWS:
		filters = append(filters, func(i versionsapi.ImageInfoEntry) bool {
			return i.Region == region
		})
	}
	return filters
}

func getFromFile(fs *afero.Afero, imgInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error) {
	fileName := imageInfoFilename(imgInfo)

	raw, err := fs.ReadFile(fileName)
	if err != nil {
		return versionsapi.ImageInfo{}, err
	}

	var newInfo versionsapi.ImageInfo
	if err := json.Unmarshal(raw, &newInfo); err != nil {
		return versionsapi.ImageInfo{}, fmt.Errorf("decoding image info file: %w", err)
	}

	return newInfo, nil
}

var filenameReplaceRegexp = regexp.MustCompile(`([^a-zA-Z0-9.-])`)

func imageInfoFilename(imgInfo versionsapi.ImageInfo) string {
	path := imgInfo.JSONPath()
	return filenameReplaceRegexp.ReplaceAllString(path, "_")
}

// getReferenceFromImageInfo returns the image reference for a given CSP and image variant.
func getReferenceFromImageInfo(provider cloudprovider.Provider,
	attestationVariant string, imgInfo versionsapi.ImageInfo,
	filt ...filter,
) (string, error) {
	var correctVariant versionsapi.ImageInfoEntry
	var found bool
variantLoop:
	for _, variant := range imgInfo.List {
		gotCSP := cloudprovider.FromString(variant.CSP)
		if gotCSP != provider || variant.AttestationVariant != attestationVariant {
			continue
		}
		for _, f := range filt {
			if !f(variant) {
				continue variantLoop
			}
		}
		correctVariant = variant
		found = true
		break
	}
	if !found {
		return "", fmt.Errorf("image not available in image info for CSP %q, variant %q and other filters", provider.String(), attestationVariant)
	}

	return correctVariant.Reference, nil
}

type versionsAPIImageInfoFetcher interface {
	FetchImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error)
}

type filter func(versionsapi.ImageInfoEntry) bool
