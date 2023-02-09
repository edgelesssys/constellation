/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package image provides helping wrappers around a versionsapi fetcher.

It also enables local image overrides and download of raw images.
*/
package image

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"regexp"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
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
		fetcher: fetcher.NewFetcher(),
		fs:      &afero.Afero{Fs: afero.NewOsFs()},
	}
}

// FetchReference fetches the image reference for a given image version uid, CSP and image variant.
func (f *Fetcher) FetchReference(ctx context.Context, config *config.Config) (string, error) {
	provider := config.GetProvider()
	variant, err := variant(provider, config)
	if err != nil {
		return "", fmt.Errorf("determining variant: %w", err)
	}

	ver, err := versionsapi.NewVersionFromShortPath(config.Image, versionsapi.VersionKindImage)
	if err != nil {
		return "", fmt.Errorf("parsing config image short path: %w", err)
	}

	imgInfoReq := versionsapi.ImageInfo{
		Ref:     ver.Ref,
		Stream:  ver.Stream,
		Version: ver.Version,
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

	return getReferenceFromImageInfo(provider, variant, imgInfo)
}

// variant returns the image variant for a given CSP and configuration.
func variant(provider cloudprovider.Provider, config *config.Config) (string, error) {
	switch provider {
	case cloudprovider.AWS:
		return config.Provider.AWS.Region, nil
	case cloudprovider.Azure:
		if *config.Provider.Azure.ConfidentialVM {
			return "cvm", nil
		}
		return "trustedlaunch", nil

	case cloudprovider.GCP:
		return "sev-es", nil
	case cloudprovider.QEMU:
		return "default", nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
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
func getReferenceFromImageInfo(provider cloudprovider.Provider, variant string, imgInfo versionsapi.ImageInfo,
) (string, error) {
	var providerList map[string]string
	switch provider {
	case cloudprovider.AWS:
		providerList = imgInfo.AWS
	case cloudprovider.Azure:
		providerList = imgInfo.Azure
	case cloudprovider.GCP:
		providerList = imgInfo.GCP
	case cloudprovider.QEMU:
		providerList = imgInfo.QEMU
	default:
		return "", fmt.Errorf("image not available in image info for CSP %q", provider.String())
	}

	if providerList == nil {
		return "", fmt.Errorf("image not available in image info for CSP %q", provider.String())
	}

	ref, ok := providerList[variant]
	if !ok {
		return "", fmt.Errorf("image not available in image info for variant %q", variant)
	}

	return ref, nil
}

type versionsAPIImageInfoFetcher interface {
	FetchImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error)
}
