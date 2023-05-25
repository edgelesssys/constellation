/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
)

// ConfigAPIFetcher fetches config API resources without authentication.
type ConfigAPIFetcher struct {
	*fetcher
}

// NewConfigAPIFetcher returns a new Fetcher.
func NewConfigAPIFetcher() *ConfigAPIFetcher {
	return &ConfigAPIFetcher{newFetcher()}
}

// NewConfigAPIFetcherWithClient returns a new Fetcher with custom http client.
func NewConfigAPIFetcherWithClient(client HTTPClient) *ConfigAPIFetcher {
	return &ConfigAPIFetcher{newFetcherWith(client)}
}

// FetchAzureSEVSNPVersionList fetches the version list information from the config API.
func (f *ConfigAPIFetcher) FetchAzureSEVSNPVersionList(ctx context.Context, attestation configapi.AzureSEVSNPVersionList) (configapi.AzureSEVSNPVersionList, error) {
	return fetch(ctx, f.httpc, attestation)
}

// FetchAzureSEVSNPVersion fetches the version information from the config API.
func (f *ConfigAPIFetcher) FetchAzureSEVSNPVersion(ctx context.Context, attestation configapi.AzureSEVSNPVersionGet) (configapi.AzureSEVSNPVersionGet, error) {
	// TODO(elchead): follow-up PR for AB#3045 to check signature (sigstore.VerifySignature)
	return fetch(ctx, f.httpc, attestation)
}

// FetchLatestAzureSEVSNPVersion returns the latest versions of the given type.
func (f *ConfigAPIFetcher) FetchLatestAzureSEVSNPVersion(ctx context.Context) (res configapi.AzureSEVSNPVersion, err error) {
	var versions configapi.AzureSEVSNPVersionList
	versions, err = f.FetchAzureSEVSNPVersionList(ctx, versions)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	get := configapi.AzureSEVSNPVersionGet{Version: versions[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = f.FetchAzureSEVSNPVersion(ctx, get)
	if err != nil {
		return res, fmt.Errorf("failed fetching version: %w", err)
	}
	return get.AzureSEVSNPVersion, nil
}
