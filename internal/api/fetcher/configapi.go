/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"
	"fmt"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

// ConfigAPIFetcher fetches config API resources without authentication.
type ConfigAPIFetcher struct {
	*fetcher
	cosignPublicKey []byte // public key to verify signatures
}

// NewConfigAPIFetcher returns a new Fetcher.
func NewConfigAPIFetcher() *ConfigAPIFetcher {
	return NewConfigAPIFetcherWithClient(NewHTTPClient())
}

// NewConfigAPIFetcherWithClient returns a new Fetcher with custom http client.
func NewConfigAPIFetcherWithClient(client HTTPClient) *ConfigAPIFetcher {
	return &ConfigAPIFetcher{
		fetcher:         newFetcherWith(client),
		cosignPublicKey: []byte(constants.CosignPublicKey), // TODO use dev key?
	}
}

// FetchAzureSEVSNPVersionList fetches the version list information from the config API.
func (f *ConfigAPIFetcher) FetchAzureSEVSNPVersionList(ctx context.Context, attestation configapi.AzureSEVSNPVersionList) (configapi.AzureSEVSNPVersionList, error) {
	return fetch(ctx, f.httpc, attestation)
}

// FetchAzureSEVSNPVersion fetches the version information from the config API.
func (f *ConfigAPIFetcher) FetchAzureSEVSNPVersion(ctx context.Context, attestation configapi.AzureSEVSNPVersionGet) (configapi.AzureSEVSNPVersionGet, error) {
	urlString, err := attestation.URL()
	if err != nil {
		return attestation, err
	}
	versionBytes, err := fetchBytesFromRawURL(ctx, urlString, f.httpc)
	if err != nil {
		return attestation, fmt.Errorf("fetch version %s: %w", attestation.Version, err)
	}

	signature, err := fetchBytesFromRawURL(ctx, fmt.Sprintf("%s.sig", urlString), f.httpc)
	if err != nil {
		return attestation, fmt.Errorf("fetch version %s signature: %w", attestation.Version, err)
	}

	err = sigstore.VerifySignature(versionBytes, signature, f.cosignPublicKey)
	if err != nil {
		return attestation, fmt.Errorf("verify version %s signature: %w", attestation.Version, err)
	}
	return fetch(ctx, f.httpc, attestation)
}

func fetchBytesFromRawURL(ctx context.Context, urlString string, client HTTPClient) ([]byte, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("parse version url %s: %w", urlString, err)
	}
	bytes, err := FetchFromURL(ctx, client, url)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// FetchLatestAzureSEVSNPVersion returns the latest versions of the given type.
func (f *ConfigAPIFetcher) FetchLatestAzureSEVSNPVersion(ctx context.Context) (res configapi.AzureSEVSNPVersion, err error) {
	var versions configapi.AzureSEVSNPVersionList
	versions, err = f.FetchAzureSEVSNPVersionList(ctx, versions)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	fmt.Println("versions", versions)
	get := configapi.AzureSEVSNPVersionGet{Version: versions[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = f.FetchAzureSEVSNPVersion(ctx, get)
	if err != nil {
		return res, fmt.Errorf("failed fetching version: %w", err)
	}
	return get.AzureSEVSNPVersion, nil
}
