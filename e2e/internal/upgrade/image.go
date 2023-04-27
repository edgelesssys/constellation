//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"gopkg.in/yaml.v3"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
)

type upgradeInfo struct {
	measurements measurements.M
	shortPath    string
	imageRef     string
}

func fetchUpgradeInfo(ctx context.Context, csp cloudprovider.Provider, toImage string) (upgradeInfo, error) {
	info := upgradeInfo{
		measurements: make(measurements.M),
		shortPath:    toImage,
	}
	versionsClient := fetcher.NewFetcher()

	ver, err := versionsapi.NewVersionFromShortPath(toImage, versionsapi.VersionKindImage)
	if err != nil {
		return upgradeInfo{}, err
	}

	measurementsURL, _, err := versionsapi.MeasurementURL(ver, csp)
	if err != nil {
		return upgradeInfo{}, err
	}

	fetchedMeasurements, err := fetchMeasurements(
		ctx, http.DefaultClient,
		measurementsURL,
		measurements.WithMetadata{
			CSP:   csp,
			Image: toImage,
		},
	)
	if err != nil {
		return upgradeInfo{}, err
	}
	info.measurements = fetchedMeasurements

	imageRef, err := fetchImageRef(ctx, versionsClient, csp, versionsapi.ImageInfo{
		Ref:     ver.Ref,
		Stream:  ver.Stream,
		Version: ver.Version,
	})
	if err != nil {
		return upgradeInfo{}, err
	}
	info.imageRef = imageRef

	return info, nil
}

// fetchMeasurements is essentially a copy of measurements.FetchAndVerify, but with verification removed.
// This is necessary since the e2e tests may target release images for which the measurements are signed with the release public key.
// It is easier to skip verification than to implement a second bazel target with the enterprise build tag set.
func fetchMeasurements(ctx context.Context, client *http.Client, measurementsURL *url.URL, metadata measurements.WithMetadata) (measurements.M, error) {
	measurementsRaw, err := getFromURL(ctx, client, measurementsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch measurements: %w", err)
	}

	var mWithMetadata measurements.WithMetadata
	if err := json.Unmarshal(measurementsRaw, &mWithMetadata); err != nil {
		if yamlErr := yaml.Unmarshal(measurementsRaw, &mWithMetadata); yamlErr != nil {
			return nil, errors.Join(
				err,
				fmt.Errorf("trying yaml format: %w", yamlErr),
			)
		}
	}

	if mWithMetadata.CSP != metadata.CSP {
		return nil, fmt.Errorf("invalid measurement metadata: CSP mismatch: expected %s, got %s", metadata.CSP, mWithMetadata.CSP)
	}
	if mWithMetadata.Image != metadata.Image {
		return nil, fmt.Errorf("invalid measurement metadata: image mismatch: expected %s, got %s", metadata.Image, mWithMetadata.Image)
	}

	return mWithMetadata.Measurements, nil
}

func getFromURL(ctx context.Context, client *http.Client, sourceURL *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL.String(), http.NoBody)
	if err != nil {
		return []byte{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

func fetchImageRef(ctx context.Context, client *fetcher.Fetcher, csp cloudprovider.Provider, imageInfo versionsapi.ImageInfo) (string, error) {
	imageInfo, err := client.FetchImageInfo(ctx, imageInfo)
	if err != nil {
		return "", err
	}

	switch csp {
	case cloudprovider.GCP:
		return imageInfo.GCP["sev-es"], nil
	case cloudprovider.Azure:
		return imageInfo.Azure["cvm"], nil
	case cloudprovider.AWS:
		return imageInfo.AWS["eu-central-1"], nil
	default:
		return "", errors.New("finding wanted image")
	}
}
