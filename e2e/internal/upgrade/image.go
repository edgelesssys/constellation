//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"errors"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
)

type upgradeInfo struct {
	measurements measurements.M
	shortPath    string
	imageRef     string
}

func fetchUpgradeInfo(ctx context.Context, csp cloudprovider.Provider,
	attestationVariant variant.Variant, toImage string,
) (upgradeInfo, error) {
	info := upgradeInfo{
		measurements: make(measurements.M),
		shortPath:    toImage,
	}
	versionsClient := fetcher.NewFetcher()

	ver, err := versionsapi.NewVersionFromShortPath(toImage, versionsapi.VersionKindImage)
	if err != nil {
		return upgradeInfo{}, err
	}

	measurementsURL, _, err := versionsapi.MeasurementURL(ver)
	if err != nil {
		return upgradeInfo{}, err
	}

	fetchedMeasurements := measurements.M{}
	if err := fetchedMeasurements.FetchNoVerify(
		ctx, http.DefaultClient,
		measurementsURL,
		ver, csp, attestationVariant,
	); err != nil {
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
