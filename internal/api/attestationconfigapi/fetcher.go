/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"context"
	"errors"
	"fmt"

	apifetcher "github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

const cosignPublicKey = constants.CosignPublicKeyReleases

// ErrNoVersionsFound is returned if no versions are found.
var ErrNoVersionsFound = errors.New("no versions found")

// Fetcher fetches config API resources without authentication.
type Fetcher interface {
	FetchAzureSEVSNPVersion(ctx context.Context, azureVersion AzureSEVSNPVersionAPI) (AzureSEVSNPVersionAPI, error)
	FetchAzureSEVSNPVersionList(ctx context.Context, attestation AzureSEVSNPVersionList) (AzureSEVSNPVersionList, error)
	FetchAzureSEVSNPVersionLatest(ctx context.Context) (AzureSEVSNPVersionAPI, error)
}

// fetcher fetches AttestationCfg API resources without authentication.
type fetcher struct {
	apifetcher.HTTPClient
	cdnURL   string
	verifier sigstore.Verifier
}

// NewFetcher returns a new apifetcher.
func NewFetcher() Fetcher {
	return NewFetcherWithClient(apifetcher.NewHTTPClient(), constants.CDNRepositoryURL)
}

// NewFetcherWithCustomCDNAndCosignKey returns a new fetcher with custom CDN URL.
func NewFetcherWithCustomCDNAndCosignKey(cdnURL, cosignKey string) Fetcher {
	verifier, err := sigstore.NewCosignVerifier([]byte(cosignKey))
	if err != nil {
		// This relies on an embedded public key. If this key can not be validated, there is no way to recover from this.
		panic(fmt.Errorf("creating cosign verifier: %w", err))
	}
	return newFetcherWithClientAndVerifier(apifetcher.NewHTTPClient(), verifier, cdnURL)
}

// NewFetcherWithClient returns a new fetcher with custom http client.
func NewFetcherWithClient(client apifetcher.HTTPClient, cdnURL string) Fetcher {
	verifier, err := sigstore.NewCosignVerifier([]byte(cosignPublicKey))
	if err != nil {
		// This relies on an embedded public key. If this key can not be validated, there is no way to recover from this.
		panic(fmt.Errorf("creating cosign verifier: %w", err))
	}
	return newFetcherWithClientAndVerifier(client, verifier, cdnURL)
}

func newFetcherWithClientAndVerifier(client apifetcher.HTTPClient, cosignVerifier sigstore.Verifier, url string) Fetcher {
	return &fetcher{HTTPClient: client, verifier: cosignVerifier, cdnURL: url}
}

// FetchAzureSEVSNPVersionList fetches the version list information from the config API.
func (f *fetcher) FetchAzureSEVSNPVersionList(ctx context.Context, attestation AzureSEVSNPVersionList) (AzureSEVSNPVersionList, error) {
	// TODO(derpsteb): Replace with FetchAndVerify once we move to v2 of the config API.
	return apifetcher.Fetch(ctx, f.HTTPClient, f.cdnURL, attestation)
}

// FetchAzureSEVSNPVersion fetches the version information from the config API.
func (f *fetcher) FetchAzureSEVSNPVersion(ctx context.Context, azureVersion AzureSEVSNPVersionAPI) (AzureSEVSNPVersionAPI, error) {
	fetchedVersion, err := apifetcher.FetchAndVerify(ctx, f.HTTPClient, f.cdnURL, azureVersion, f.verifier)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetching version %s: %w", azureVersion.Version, err)
	}
	return fetchedVersion, nil
}

// FetchAzureSEVSNPVersionLatest returns the latest versions of the given type.
func (f *fetcher) FetchAzureSEVSNPVersionLatest(ctx context.Context) (res AzureSEVSNPVersionAPI, err error) {
	var list AzureSEVSNPVersionList
	list, err = f.FetchAzureSEVSNPVersionList(ctx, list)
	if err != nil {
		return res, ErrNoVersionsFound
	}
	if len(list) < 1 {
		return res, ErrNoVersionsFound
	}
	getVersionRequest := AzureSEVSNPVersionAPI{
		Version: list[0], // latest version is first in list
	}
	res, err = f.FetchAzureSEVSNPVersion(ctx, getVersionRequest)
	if err != nil {
		return res, err
	}
	return
}
