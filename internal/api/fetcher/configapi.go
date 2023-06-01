/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

// ConfigAPIFetcher fetches config API resources without authentication.
type ConfigAPIFetcher interface {
	FetchAzureSEVSNPVersionList(ctx context.Context, attestation configapi.AzureSEVSNPVersionList) (configapi.AzureSEVSNPVersionList, error)
	FetchAzureSEVSNPVersion(ctx context.Context, azureVersion configapi.AzureSEVSNPVersionGet, version versionsapi.Version) (configapi.AzureSEVSNPVersionGet, error)
	FetchLatestAzureSEVSNPVersion(ctx context.Context, version versionsapi.Version) (configapi.AzureSEVSNPVersion, error)
}

// configAPIFetcher fetches config API resources without authentication.
type configAPIFetcher struct {
	*fetcher
}

// NewConfigAPIFetcher returns a new Fetcher.
func NewConfigAPIFetcher() ConfigAPIFetcher {
	return NewConfigAPIFetcherWithClient(NewHTTPClient())
}

// NewConfigAPIFetcherWithClient returns a new Fetcher with custom http client.
func NewConfigAPIFetcherWithClient(client HTTPClient) ConfigAPIFetcher {
	return &configAPIFetcher{
		fetcher: newFetcherWith(client),
	}
}

// FetchAzureSEVSNPVersionList fetches the version list information from the config API.
func (f *configAPIFetcher) FetchAzureSEVSNPVersionList(ctx context.Context, attestation configapi.AzureSEVSNPVersionList) (configapi.AzureSEVSNPVersionList, error) {
	return fetch(ctx, f.httpc, attestation)
}

// FetchAzureSEVSNPVersion fetches the version information from the config API.
func (f *configAPIFetcher) FetchAzureSEVSNPVersion(ctx context.Context, azureVersion configapi.AzureSEVSNPVersionGet, version versionsapi.Version) (configapi.AzureSEVSNPVersionGet, error) {
	cosignPublicKey, err := sigstore.CosignPublicKeyForVersion(version)
	if err != nil {
		return azureVersion, fmt.Errorf("get public key for config: %w", err)
	}
	urlString, err := azureVersion.URL()
	if err != nil {
		return azureVersion, err
	}
	fetchedVersion, err := fetch(ctx, f.httpc, azureVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s: %w", fetchedVersion.Version, err)
	}
	versionBytes, err := json.Marshal(fetchedVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("marshal version for verify %s: %w", fetchedVersion.Version, err)
	}

	signature, err := fetchBytesFromRawURL(ctx, fmt.Sprintf("%s.sig", urlString), f.httpc)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s signature: %w", fetchedVersion.Version, err)
	}

	err = sigstore.CosignVerifier{}.VerifySignature(versionBytes, signature, cosignPublicKey)
	if err != nil {
		return fetchedVersion, fmt.Errorf("verify version %s signature: %w", fetchedVersion.Version, err)
	}
	return fetchedVersion, nil
}

// FetchLatestAzureSEVSNPVersion returns the latest versions of the given type.
func (f *configAPIFetcher) FetchLatestAzureSEVSNPVersion(ctx context.Context, version versionsapi.Version) (res configapi.AzureSEVSNPVersion, err error) {
	var versions configapi.AzureSEVSNPVersionList
	versions, err = f.FetchAzureSEVSNPVersionList(ctx, versions)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	get := configapi.AzureSEVSNPVersionGet{Version: versions[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = f.FetchAzureSEVSNPVersion(ctx, get, version)
	if err != nil {
		return res, fmt.Errorf("failed fetching version: %w", err)
	}
	return get.AzureSEVSNPVersion, nil
}

func fetchBytesFromRawURL(ctx context.Context, urlString string, client HTTPClient) ([]byte, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("parse version url %s: %w", urlString, err)
	}
	return getFromURL(ctx, client, url)
}
