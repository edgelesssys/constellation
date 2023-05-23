//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

type upgradeInfo struct {
	measurements measurements.M
	shortPath    string
	imageRef     string
}

func fetchUpgradeInfo(ctx context.Context, csp cloudprovider.Provider,
	attestationVariant variant.Variant, toImage, region string,
) (upgradeInfo, error) {
	info := upgradeInfo{
		measurements: make(measurements.M),
		shortPath:    toImage,
	}

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

	fetcher := imagefetcher.New()
	imageRef, err := fetcher.FetchReference(ctx, csp, attestationVariant, toImage, region)
	if err != nil {
		return upgradeInfo{}, err
	}
	info.imageRef = imageRef

	return info, nil
}
