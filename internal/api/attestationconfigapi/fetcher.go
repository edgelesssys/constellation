/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apifetcher "github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

// minimumAgeVersion is the minimum age to accept the version as latest.
const minimumAgeVersion = 14 * 24 * time.Hour

const cosignPublicKey = constants.CosignPublicKeyReleases

// Fetcher fetches config API resources without authentication.
type Fetcher interface {
	FetchAzureSEVSNPVersion(ctx context.Context, azureVersion AzureSEVSNPVersionAPI) (AzureSEVSNPVersionAPI, error)
	FetchAzureSEVSNPVersionList(ctx context.Context, attestation AzureSEVSNPVersionList) (AzureSEVSNPVersionList, error)
	FetchAzureSEVSNPVersionLatest(ctx context.Context, now time.Time) (AzureSEVSNPVersionAPI, error)
}

// fetcher fetches AttestationCfg API resources without authentication.
type fetcher struct {
	apifetcher.HTTPClient
	verifier sigstore.Verifier
}

// NewFetcher returns a new apifetcher.
func NewFetcher() Fetcher {
	return NewFetcherWithClient(apifetcher.NewHTTPClient())
}

// NewFetcherWithClient returns a new fetcher with custom http client.
func NewFetcherWithClient(client apifetcher.HTTPClient) Fetcher {
	return newFetcherWithClientAndVerifier(client, sigstore.CosignVerifier{})
}

func newFetcherWithClientAndVerifier(client apifetcher.HTTPClient, cosignVerifier sigstore.Verifier) Fetcher {
	return &fetcher{client, cosignVerifier}
}

// FetchAzureSEVSNPVersionList fetches the version list information from the config API.
func (f *fetcher) FetchAzureSEVSNPVersionList(ctx context.Context, attestation AzureSEVSNPVersionList) (AzureSEVSNPVersionList, error) {
	return apifetcher.Fetch(ctx, f.HTTPClient, attestation)
}

// FetchAzureSEVSNPVersion fetches the version information from the config API.
func (f *fetcher) FetchAzureSEVSNPVersion(ctx context.Context, azureVersion AzureSEVSNPVersionAPI) (AzureSEVSNPVersionAPI, error) {
	fetchedVersion, err := apifetcher.Fetch(ctx, f.HTTPClient, azureVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s: %w", fetchedVersion.Version, err)
	}
	versionBytes, err := json.Marshal(fetchedVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("marshal version for verify %s: %w", azureVersion.Version, err)
	}

	signature, err := apifetcher.Fetch(ctx, f.HTTPClient, AzureSEVSNPVersionSignature{
		Version: azureVersion.Version,
	})
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s signature: %w", azureVersion.Version, err)
	}

	err = f.verifier.VerifySignature(versionBytes, signature.Signature, []byte(cosignPublicKey))
	if err != nil {
		return fetchedVersion, fmt.Errorf("verify version %s signature: %w", azureVersion.Version, err)
	}
	return fetchedVersion, nil
}

// FetchAzureSEVSNPVersionLatest returns the latest versions of the given type.
func (f *fetcher) FetchAzureSEVSNPVersionLatest(ctx context.Context, now time.Time) (res AzureSEVSNPVersionAPI, err error) {
	var list AzureSEVSNPVersionList
	list, err = f.FetchAzureSEVSNPVersionList(ctx, list)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	getVersionRequest, err := getLatestVersionOlderThanMinimumAge(list, now, minimumAgeVersion)
	if err != nil {
		return res, fmt.Errorf("finding latest valid version: %w", err)
	}
	res, err = f.FetchAzureSEVSNPVersion(ctx, getVersionRequest)
	if err != nil {
		return res, fmt.Errorf("fetching version: %w", err)
	}
	return
}

func getLatestVersionOlderThanMinimumAge(list AzureSEVSNPVersionList, now time.Time, minimumAgeVersion time.Duration) (AzureSEVSNPVersionAPI, error) {
	SortAzureSEVSNPVersionList(list)
	for _, v := range list {
		dateStr := strings.TrimSuffix(v, ".json")
		versionDate, err := time.Parse("2006-01-01-01-01", dateStr)
		if err != nil {
			return AzureSEVSNPVersionAPI{}, fmt.Errorf("parsing version date %s: %w", dateStr, err)
		}
		if now.Sub(versionDate) > minimumAgeVersion {
			return AzureSEVSNPVersionAPI{Version: v}, nil
		}
	}
	return AzureSEVSNPVersionAPI{}, fmt.Errorf("no valid version fulfilling minimum age found")
}
