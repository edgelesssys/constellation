/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"context"
	"fmt"

	apifetcher "github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

const cosignPublicKey = constants.CosignPublicKeyReleases

// Fetcher fetches config API resources without authentication.
type Fetcher interface {
	FetchLatestVersion(ctx context.Context, attestation variant.Variant) (Entry, error)
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

// FetchLatestVersion returns the latest versions of the given type.
func (f *fetcher) FetchLatestVersion(ctx context.Context, variant variant.Variant) (Entry, error) {
	list, err := f.fetchVersionList(ctx, variant)
	if err != nil {
		return Entry{}, err
	}

	// latest version is first in list
	return f.fetchVersion(ctx, list.List[0], variant)
}

// fetchVersionList fetches the version list information from the config API.
func (f *fetcher) fetchVersionList(ctx context.Context, variant variant.Variant) (List, error) {
	fetchedList, err := apifetcher.FetchAndVerify(ctx, f.HTTPClient, f.cdnURL, List{Variant: variant}, f.verifier)
	if err != nil {
		return List{}, fmt.Errorf("fetching version list: %w", err)
	}

	// Set the attestation variant of the list as it is not part of the marshalled JSON retrieved by Fetch
	fetchedList.Variant = variant

	return fetchedList, nil
}

// fetchVersion fetches the version information from the config API.
func (f *fetcher) fetchVersion(ctx context.Context, version string, variant variant.Variant) (Entry, error) {
	obj := Entry{
		Version: version,
		Variant: variant,
	}
	fetchedVersion, err := apifetcher.FetchAndVerify(ctx, f.HTTPClient, f.cdnURL, obj, f.verifier)
	if err != nil {
		return Entry{}, fmt.Errorf("fetching version %q: %w", version, err)
	}

	// Set the attestation variant of the list as it is not part of the marshalled JSON retrieved by FetchAndVerify
	fetchedVersion.Variant = variant

	return fetchedVersion, nil
}
