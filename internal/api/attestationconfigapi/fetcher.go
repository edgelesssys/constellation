/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"context"
	"encoding/json"
	"fmt"

	apifetcher "github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

const cosignPublicKey = constants.CosignPublicKeyReleases

// Fetcher fetches config API resources without authentication.
type Fetcher interface {
	FetchAzureSEVSNPVersion(ctx context.Context, azureVersion AzureSEVSNPVersionAPI) (AzureSEVSNPVersionAPI, error)
	FetchAzureSEVSNPVersionList(ctx context.Context, attestation AzureSEVSNPVersionList) (AzureSEVSNPVersionList, error)
	FetchAzureSEVSNPVersionLatest(ctx context.Context) (AzureSEVSNPVersionAPI, error)
}

// fetcher fetches AttestationCfg API resources without authentication.
type fetcher struct {
	apifetcher.HTTPClient
}

// NewFetcher returns a new apifetcher.
func NewFetcher() Fetcher {
	return NewFetcherWithClient(apifetcher.NewHTTPClient())
}

// NewFetcherWithClient returns a new fetcher with custom http client.
func NewFetcherWithClient(client apifetcher.HTTPClient) Fetcher {
	return &fetcher{client}
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

	err = sigstore.CosignVerifier{}.VerifySignature(versionBytes, signature.Signature, []byte(cosignPublicKey))
	if err != nil {
		return fetchedVersion, fmt.Errorf("verify version %s signature: %w", azureVersion.Version, err)
	}
	return fetchedVersion, nil
}

// FetchAzureSEVSNPVersionLatest returns the latest versions of the given type.
func (f *fetcher) FetchAzureSEVSNPVersionLatest(ctx context.Context) (res AzureSEVSNPVersionAPI, err error) {
	var list AzureSEVSNPVersionList
	list, err = f.FetchAzureSEVSNPVersionList(ctx, list)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	get := AzureSEVSNPVersionAPI{Version: list[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = f.FetchAzureSEVSNPVersion(ctx, get)
	if err != nil {
		return res, fmt.Errorf("fetching version: %w", err)
	}
	return get, nil
}
